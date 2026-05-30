package todex

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goodway/godos/internal/todexapi"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

var (
	ErrListNotFound     = errors.New("list not found")
	ErrListExists       = errors.New("list already exists")
	ErrTaskNotFound     = errors.New("task not found")
	ErrAmbiguousID      = errors.New("ambiguous task ID prefix")
	ErrResponseTooLarge = errors.New("Todex API response too large")
)

const (
	defaultHTTPTimeout = time.Second
	maxResponseBytes   = 1024 * 1024
)

type ServiceConfig struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

type Service struct {
	client *todexapi.ClientWithResponses
	tasks  []Task
}

type ListSummary struct {
	Name      string
	Completed int
	Total     int
}

type ListTasks struct {
	Name  string
	Tasks []Task
}

type Task struct {
	ID      string
	ShortID string
	Title   string
	Done    bool
}

type User struct {
	Email string
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	resp, err := s.client.LoginUserWithResponse(ctx, todexapi.LoginRequest{Email: openapi_types.Email(email), Password: password})
	if err != nil {
		return "", normalizeTransportError(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Data.Token == nil || *resp.JSON200.Data.Token == "" {
		return "", normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return *resp.JSON200.Data.Token, nil
}

func (s *Service) Register(ctx context.Context, email, password string) (string, error) {
	resp, err := s.client.RegisterUserWithResponse(ctx, todexapi.RegisterRequest{Email: openapi_types.Email(email), Password: password})
	if err != nil {
		return "", normalizeTransportError(err)
	}
	if resp.JSON201 == nil || resp.JSON201.Data.Token == nil || *resp.JSON201.Data.Token == "" {
		return "", normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return *resp.JSON201.Data.Token, nil
}

func (s *Service) CurrentUser(ctx context.Context) (User, error) {
	resp, err := s.client.GetAuthMeWithResponse(ctx)
	if err != nil {
		return User{}, normalizeTransportError(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Data.User == nil {
		return User{}, normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return User{Email: string(resp.JSON200.Data.User.Email)}, nil
}

func (s *Service) Logout(ctx context.Context) error {
	resp, err := s.client.LogoutUserWithResponse(ctx)
	if err != nil {
		return normalizeTransportError(err)
	}
	if resp.JSON200 == nil {
		return normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return nil
}

func New(cfg ServiceConfig) (*Service, error) {
	baseURL := strings.TrimSpace(cfg.BaseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("Todex API base URL is not configured; run godos configure set api_base_url <url> or set GODOS_API_BASE_URL")
	}
	if err := validateBaseURL(baseURL); err != nil {
		return nil, err
	}

	options := []todexapi.ClientOption{todexapi.WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
		if cfg.Token != "" {
			req.Header.Set("Authorization", "Bearer "+cfg.Token)
		}
		return nil
	})}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultHTTPTimeout, CheckRedirect: safeRedirect}
	}
	options = append(options, todexapi.WithHTTPClient(limitResponseDoer{next: httpClient, max: maxResponseBytes}))

	client, err := todexapi.NewClientWithResponses(baseURL, options...)
	if err != nil {
		return nil, err
	}
	return &Service{client: client}, nil
}

func (s *Service) ListSummaries(ctx context.Context) ([]ListSummary, error) {
	listTasks, err := s.ListAllTasks(ctx)
	if err != nil {
		return nil, err
	}

	summaries := make([]ListSummary, 0, len(listTasks))
	for _, list := range listTasks {
		completed := 0
		for _, task := range list.Tasks {
			if task.Done {
				completed++
			}
		}
		summaries = append(summaries, ListSummary{Name: list.Name, Completed: completed, Total: len(list.Tasks)})
	}
	return summaries, nil
}

func (s *Service) ListAllTasks(ctx context.Context) ([]ListTasks, error) {
	lists, err := s.lists(ctx)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.ListTasksWithResponse(ctx, nil)
	if err != nil {
		return nil, normalizeTransportError(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Data.Tasks == nil {
		return nil, normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}

	byListID := make(map[string][]Task, len(lists))
	for _, task := range *resp.JSON200.Data.Tasks {
		id := task.ListId.String()
		byListID[id] = append(byListID[id], toTask(task))
	}

	result := make([]ListTasks, 0, len(lists))
	for _, list := range lists {
		result = append(result, ListTasks{Name: list.Name, Tasks: byListID[list.Id.String()]})
	}
	return result, nil
}

func (s *Service) ListTasks(ctx context.Context, listName string) ([]Task, error) {
	list, err := s.resolveList(ctx, listName)
	if err != nil {
		return nil, err
	}
	tasks, err := s.tasksForList(ctx, list.Id)
	if err != nil {
		return nil, err
	}
	return toTasks(tasks), nil
}

func (s *Service) AddTask(ctx context.Context, listName, title string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, fmt.Errorf("todo text cannot be empty")
	}

	list, err := s.resolveList(ctx, listName)
	if errors.Is(err, ErrListNotFound) {
		list, err = s.CreateList(ctx, listName)
	}
	if err != nil {
		return Task{}, err
	}

	req := todexapi.TaskRequest{ListId: &list.Id, Title: &title}
	resp, err := s.client.CreateTaskWithResponse(ctx, req)
	if err != nil {
		return Task{}, normalizeTransportError(err)
	}
	if resp.JSON201 == nil || resp.JSON201.Data.Task == nil {
		return Task{}, normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return toTask(*resp.JSON201.Data.Task), nil
}

func (s *Service) CreateList(ctx context.Context, name string) (todexapi.List, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return todexapi.List{}, fmt.Errorf("list name cannot be empty")
	}
	req := todexapi.ListRequest{Name: &name}
	resp, err := s.client.CreateListWithResponse(ctx, req)
	if err != nil {
		return todexapi.List{}, normalizeTransportError(err)
	}
	if resp.JSON201 == nil || resp.JSON201.Data.List == nil {
		return todexapi.List{}, normalizeListResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return *resp.JSON201.Data.List, nil
}

func (s *Service) RenameList(ctx context.Context, oldName, newName string) error {
	list, err := s.resolveList(ctx, oldName)
	if err != nil {
		return err
	}
	req := todexapi.ListRequest{Name: &newName}
	resp, err := s.client.UpdateListWithResponse(ctx, list.Id, req)
	if err != nil {
		return normalizeTransportError(err)
	}
	if resp.JSON200 == nil {
		return normalizeListResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return nil
}

func (s *Service) DeleteList(ctx context.Context, name string) error {
	list, err := s.resolveList(ctx, name)
	if err != nil {
		return err
	}
	resp, err := s.client.DeleteListWithResponse(ctx, list.Id)
	if err != nil {
		return normalizeTransportError(err)
	}
	if resp.JSON200 == nil {
		return normalizeListResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return nil
}

func (s *Service) CompleteTask(ctx context.Context, prefix string) (Task, error) {
	id, err := s.resolveRemoteTaskPrefix(ctx, prefix)
	if err != nil {
		return Task{}, err
	}
	resp, err := s.client.CompleteTaskWithResponse(ctx, uuid.MustParse(id))
	if err != nil {
		return Task{}, normalizeTransportError(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Data.Task == nil {
		return Task{}, normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return toTask(*resp.JSON200.Data.Task), nil
}

func (s *Service) ReopenTask(ctx context.Context, prefix string) (Task, error) {
	id, err := s.resolveRemoteTaskPrefix(ctx, prefix)
	if err != nil {
		return Task{}, err
	}
	resp, err := s.client.ReopenTaskWithResponse(ctx, uuid.MustParse(id))
	if err != nil {
		return Task{}, normalizeTransportError(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Data.Task == nil {
		return Task{}, normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return toTask(*resp.JSON200.Data.Task), nil
}

func (s *Service) DeleteTask(ctx context.Context, prefix string) (Task, error) {
	id, err := s.resolveRemoteTaskPrefix(ctx, prefix)
	if err != nil {
		return Task{}, err
	}
	resp, err := s.client.DeleteTaskWithResponse(ctx, uuid.MustParse(id))
	if err != nil {
		return Task{}, normalizeTransportError(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Data.Task == nil {
		return Task{}, normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return toTask(*resp.JSON200.Data.Task), nil
}

func (s *Service) ResolveTaskPrefix(prefix string) (string, error) {
	return resolveTaskPrefix(s.tasks, prefix)
}

func (s *Service) resolveRemoteTaskPrefix(ctx context.Context, prefix string) (string, error) {
	resp, err := s.client.ListTasksWithResponse(ctx, nil)
	if err != nil {
		return "", normalizeTransportError(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Data.Tasks == nil {
		return "", normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return resolveTaskPrefix(toTasks(*resp.JSON200.Data.Tasks), prefix)
}

func resolveTaskPrefix(tasks []Task, prefix string) (string, error) {
	prefix = strings.TrimSpace(strings.ToLower(prefix))
	if prefix == "" {
		return "", fmt.Errorf("task ID prefix is required")
	}
	if _, err := strconv.Atoi(prefix); err == nil {
		return "", fmt.Errorf("%q is a positional number; provide a task ID prefix", prefix)
	}

	matches := make([]Task, 0, 1)
	for _, task := range tasks {
		if strings.HasPrefix(strings.ToLower(task.ID), prefix) {
			matches = append(matches, task)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("%w: %q", ErrTaskNotFound, prefix)
	case 1:
		return matches[0].ID, nil
	default:
		return "", fmt.Errorf("%w %q; provide more characters", ErrAmbiguousID, prefix)
	}
}

func (s *Service) resolveList(ctx context.Context, name string) (todexapi.List, error) {
	lists, err := s.lists(ctx)
	if err != nil {
		return todexapi.List{}, err
	}
	matches := make([]todexapi.List, 0, 1)
	for _, list := range lists {
		if list.Name == name {
			matches = append(matches, list)
		}
	}
	switch len(matches) {
	case 0:
		return todexapi.List{}, fmt.Errorf("%w: %q", ErrListNotFound, name)
	case 1:
		return matches[0], nil
	default:
		return todexapi.List{}, fmt.Errorf("duplicate remote list name %q", name)
	}
}

func (s *Service) lists(ctx context.Context) ([]todexapi.List, error) {
	resp, err := s.client.ListListsWithResponse(ctx)
	if err != nil {
		return nil, normalizeTransportError(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Data.Lists == nil {
		return nil, normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return *resp.JSON200.Data.Lists, nil
}

func (s *Service) tasksForList(ctx context.Context, listID openapi_types.UUID) ([]todexapi.Task, error) {
	params := &todexapi.ListTasksParams{ListId: &listID}
	resp, err := s.client.ListTasksWithResponse(ctx, params)
	if err != nil {
		return nil, normalizeTransportError(err)
	}
	if resp.JSON200 == nil || resp.JSON200.Data.Tasks == nil {
		return nil, normalizeResponse(resp.StatusCode(), resp.Body, firstError(resp.JSON400, resp.JSON401, resp.JSON404, resp.JSON415, resp.JSON422))
	}
	return *resp.JSON200.Data.Tasks, nil
}

func toTasks(apiTasks []todexapi.Task) []Task {
	tasks := make([]Task, 0, len(apiTasks))
	for _, task := range apiTasks {
		tasks = append(tasks, toTask(task))
	}
	return tasks
}

func toTask(task todexapi.Task) Task {
	id := task.Id.String()
	shortID := id
	if len(shortID) > 8 {
		shortID = shortID[:8]
	}
	return Task{ID: id, ShortID: shortID, Title: task.Title, Done: task.Status == todexapi.Completed}
}

func normalizeTransportError(err error) error {
	if errors.Is(err, ErrResponseTooLarge) {
		return ErrResponseTooLarge
	}
	if errors.Is(err, io.ErrUnexpectedEOF) || strings.Contains(err.Error(), "unexpected end of JSON input") {
		return fmt.Errorf("malformed response from Todex API: %w", err)
	}
	return fmt.Errorf("connecting to Todex API: %w", err)
}

func normalizeListResponse(status int, body []byte, apiErr *todexapi.ErrorResponse) error {
	err := normalizeResponse(status, body, apiErr)
	if status == http.StatusConflict || strings.Contains(err.Error(), "already exists") {
		return fmt.Errorf("%w: %v", ErrListExists, err)
	}
	if status == http.StatusNotFound || strings.Contains(err.Error(), "not found") {
		return fmt.Errorf("%w: %v", ErrListNotFound, err)
	}
	return err
}

func validateBaseURL(value string) error {
	u, err := url.ParseRequestURI(value)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("Todex API base URL must be a valid absolute URL")
	}
	if u.User != nil {
		return fmt.Errorf("Todex API base URL must not include credentials")
	}
	if u.Scheme == "https" || (u.Scheme == "http" && isLoopbackHost(u.Hostname())) {
		return nil
	}
	return fmt.Errorf("Todex API base URL must use https, except for localhost development")
}

func isLoopbackHost(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func safeRedirect(req *http.Request, via []*http.Request) error {
	if len(via) == 0 {
		return nil
	}
	previous := via[len(via)-1].URL
	if req.URL.User != nil {
		return fmt.Errorf("refusing Todex API redirect with credentials")
	}
	if previous.Scheme == "https" && req.URL.Scheme != "https" {
		return fmt.Errorf("refusing Todex API redirect from https to %s", req.URL.Scheme)
	}
	if !strings.EqualFold(previous.Hostname(), req.URL.Hostname()) {
		return fmt.Errorf("refusing Todex API redirect to different host")
	}
	return nil
}

type limitResponseDoer struct {
	next todexapi.HttpRequestDoer
	max  int64
}

func (d limitResponseDoer) Do(req *http.Request) (*http.Response, error) {
	resp, err := d.next.Do(req)
	if err != nil || resp == nil || resp.Body == nil {
		return resp, err
	}
	resp.Body = &limitedReadCloser{reader: resp.Body, closer: resp.Body, remaining: d.max}
	return resp, nil
}

type limitedReadCloser struct {
	reader    io.Reader
	closer    io.Closer
	remaining int64
}

func (r *limitedReadCloser) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		buf := make([]byte, 1)
		n, err := r.reader.Read(buf)
		if n > 0 {
			return 0, ErrResponseTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.reader.Read(p)
	r.remaining -= int64(n)
	return n, err
}

func (r *limitedReadCloser) Close() error {
	return r.closer.Close()
}

func normalizeResponse(status int, body []byte, apiErr *todexapi.ErrorResponse) error {
	if apiErr != nil && apiErr.Error.Message != "" {
		return errors.New(apiErr.Error.Message)
	}
	if status == http.StatusUnauthorized {
		return errors.New("authentication required or expired; run godos login")
	}
	var decoded todexapi.ErrorResponse
	if len(body) > 0 && json.Unmarshal(body, &decoded) == nil && decoded.Error.Message != "" {
		return errors.New(decoded.Error.Message)
	}
	if status == http.StatusNotFound {
		return errors.New("Todex resource not found")
	}
	if status >= 400 {
		return fmt.Errorf("Todex API returned HTTP %d", status)
	}
	return errors.New("Todex API returned an unexpected response")
}

func firstError(errors ...*todexapi.ErrorResponse) *todexapi.ErrorResponse {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}
