package ui

import (
	"crypto/subtle"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Actual-Outcomes/doit/internal/auth"
	"github.com/Actual-Outcomes/doit/internal/model"
	"github.com/Actual-Outcomes/doit/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// UIHandlers serves the web UI pages.
type UIHandlers struct {
	store         store.Store
	signingKey    string
	adminKey      string
	adminTenantID *uuid.UUID
	templates     map[string]*template.Template
}

// NewUIHandlers parses templates once and returns a handler set.
func NewUIHandlers(s store.Store, signingKey, adminKey string, adminTenantID *uuid.UUID) *UIHandlers {
	h := &UIHandlers{
		store:         s,
		signingKey:    signingKey,
		adminKey:      adminKey,
		adminTenantID: adminTenantID,
		templates:     make(map[string]*template.Template),
	}

	pages := map[string]string{
		"login":       loginPage,
		"dashboard":   dashboardPage,
		"issues":      issuesPage,
		"issueDetail": issueDetailPage,
		"ready":       readyPage,
		"error":       errorPage,
	}

	base := template.Must(template.New("base").Funcs(templateFuncs).Parse(baseLayout))

	for name, content := range pages {
		t := template.Must(base.Clone())
		template.Must(t.Parse(content))
		h.templates[name] = t
	}

	return h
}

func (h *UIHandlers) render(w http.ResponseWriter, name string, data any) {
	t, ok := h.templates[name]
	if !ok {
		http.Error(w, "template not found", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, data); err != nil {
		slog.Error("template render error", "template", name, "error", err)
	}
}

func (h *UIHandlers) renderError(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	h.render(w, "error", map[string]any{
		"Title":     "Error",
		"ShowNav":   true,
		"NavActive": "",
		"Code":      code,
		"Message":   message,
	})
}

// LoginPage shows the login form.
func (h *UIHandlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if _, err := verifyCookie(cookie.Value, h.signingKey); err == nil {
			http.Redirect(w, r, "/ui/", http.StatusFound)
			return
		}
	}

	h.render(w, "login", map[string]any{
		"Title":   "Login",
		"ShowNav": false,
		"Error":   "",
	})
}

// LoginSubmit validates the API key and creates a session.
func (h *UIHandlers) LoginSubmit(w http.ResponseWriter, r *http.Request) {
	apiKey := r.FormValue("api_key")
	if apiKey == "" {
		h.render(w, "login", map[string]any{
			"Title":   "Login",
			"ShowNav": false,
			"Error":   "API key is required.",
		})
		return
	}

	// Check admin key first
	if h.adminKey != "" && subtle.ConstantTimeCompare([]byte(apiKey), []byte(h.adminKey)) == 1 {
		if h.adminTenantID != nil {
			setSessionCookie(w, *h.adminTenantID, h.signingKey)
			http.Redirect(w, r, "/ui/", http.StatusFound)
			return
		}
	}

	// Try resolving as tenant API key
	hash := auth.HashKey(apiKey)
	tenantID, err := h.store.ResolveAPIKey(r.Context(), hash)
	if err != nil {
		h.render(w, "login", map[string]any{
			"Title":   "Login",
			"ShowNav": false,
			"Error":   "Invalid API key.",
		})
		return
	}

	setSessionCookie(w, tenantID, h.signingKey)
	http.Redirect(w, r, "/ui/", http.StatusFound)
}

// Logout clears the session and redirects to login.
func (h *UIHandlers) Logout(w http.ResponseWriter, r *http.Request) {
	clearSessionCookie(w)
	http.Redirect(w, r, "/ui/login", http.StatusFound)
}

// addProjectData loads projects and current selection into template data.
// It uses a clean context (without project filter) to list all projects.
func (h *UIHandlers) addProjectData(r *http.Request, data map[string]any) {
	// Build a context without the project filter so we list all projects
	ctx := auth.WithTenant(r.Context(), h.tenantIDFromRequest(r))
	projects, err := h.store.ListProjects(ctx)
	if err != nil {
		slog.Error("addProjectData: list projects failed", "error", err)
		projects = nil
	}
	data["Projects"] = projects
	data["CurrentProject"] = getProjectCookie(r)
}

// tenantIDFromRequest extracts the tenant ID from the session cookie.
func (h *UIHandlers) tenantIDFromRequest(r *http.Request) uuid.UUID {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return uuid.Nil
	}
	sess, err := verifyCookie(cookie.Value, h.signingKey)
	if err != nil {
		return uuid.Nil
	}
	return sess.TenantID
}

// ProjectSwitch handles POST /ui/project to switch the active project.
func (h *UIHandlers) ProjectSwitch(w http.ResponseWriter, r *http.Request) {
	projectID := r.FormValue("project_id")
	if projectID == "" {
		clearProjectCookie(w)
	} else {
		setProjectCookie(w, projectID)
	}

	// Redirect back to referrer or dashboard
	ref := r.Header.Get("Referer")
	if ref == "" {
		ref = "/ui/"
	}
	http.Redirect(w, r, ref, http.StatusFound)
}

// Dashboard shows the overview page with stat cards.
func (h *UIHandlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	counts, err := h.store.CountIssuesByStatus(r.Context())
	if err != nil {
		slog.Error("dashboard: count query failed", "error", err)
		h.renderError(w, http.StatusInternalServerError, "Failed to load dashboard data.")
		return
	}

	// Get recent issues
	recent, err := h.store.ListIssues(r.Context(), model.IssueFilter{
		Limit:  10,
		SortBy: "updated",
	})
	if err != nil {
		slog.Error("dashboard: recent issues query failed", "error", err)
		recent = nil
	}

	// Get ready work
	ready, err := h.store.ListReady(r.Context(), model.IssueFilter{Limit: 5})
	if err != nil {
		slog.Error("dashboard: ready query failed", "error", err)
		ready = nil
	}

	data := map[string]any{
		"Title":     "Dashboard",
		"ShowNav":   true,
		"NavActive": "dashboard",
		"Counts":    counts,
		"Recent":    recent,
		"Ready":     ready,
	}
	h.addProjectData(r, data)
	h.render(w, "dashboard", data)
}

// IssueList lists issues with filtering.
func (h *UIHandlers) IssueList(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.IssueFilter{
		Limit:  50,
		SortBy: "updated",
	}

	if s := q.Get("status"); s != "" {
		status := model.Status(s)
		filter.Status = &status
	}
	if t := q.Get("type"); t != "" {
		issueType := model.IssueType(t)
		filter.IssueType = &issueType
	}
	if p := q.Get("priority"); p != "" {
		if prio, err := strconv.Atoi(p); err == nil {
			filter.Priority = &prio
		}
	}
	if a := q.Get("assignee"); a != "" {
		filter.Assignee = &a
	}

	issues, err := h.store.ListIssues(r.Context(), filter)
	if err != nil {
		slog.Error("issues: list query failed", "error", err)
		h.renderError(w, http.StatusInternalServerError, "Failed to load issues.")
		return
	}

	data := map[string]any{
		"Title":          "Issues",
		"ShowNav":        true,
		"NavActive":      "issues",
		"Issues":         issues,
		"FilterStatus":   q.Get("status"),
		"FilterType":     q.Get("type"),
		"FilterPriority": q.Get("priority"),
		"FilterAssignee": q.Get("assignee"),
	}
	h.addProjectData(r, data)
	h.render(w, "issues", data)
}

// IssueDetail shows a single issue with labels, deps, and comments.
func (h *UIHandlers) IssueDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	issue, err := h.store.GetIssue(r.Context(), id)
	if err != nil || issue == nil {
		h.renderError(w, http.StatusNotFound, "Issue not found.")
		return
	}

	comments, err := h.store.ListComments(r.Context(), id)
	if err != nil {
		slog.Error("issue detail: comments query failed", "error", err)
		comments = nil
	}

	deps, err := h.store.ListDependencies(r.Context(), id, "both")
	if err != nil {
		slog.Error("issue detail: deps query failed", "error", err)
		deps = nil
	}

	data := map[string]any{
		"Title":        issue.Title,
		"ShowNav":      true,
		"NavActive":    "issues",
		"Issue":        issue,
		"Comments":     comments,
		"Dependencies": deps,
	}
	h.addProjectData(r, data)
	h.render(w, "issueDetail", data)
}

// ReadyWork shows issues ready for work.
func (h *UIHandlers) ReadyWork(w http.ResponseWriter, r *http.Request) {
	ready, err := h.store.ListReady(r.Context(), model.IssueFilter{Limit: 50})
	if err != nil {
		slog.Error("ready: query failed", "error", err)
		h.renderError(w, http.StatusInternalServerError, "Failed to load ready work.")
		return
	}

	data := map[string]any{
		"Title":     "Ready Work",
		"ShowNav":   true,
		"NavActive": "ready",
		"Issues":    ready,
	}
	h.addProjectData(r, data)
	h.render(w, "ready", data)
}

// priorityLabel returns a human-readable priority label.
func priorityLabel(p int) string {
	switch p {
	case 0:
		return "P0 Critical"
	case 1:
		return "P1 High"
	case 2:
		return "P2 Medium"
	case 3:
		return "P3 Low"
	case 4:
		return "P4 Backlog"
	default:
		return "P" + strconv.Itoa(p)
	}
}

// statusClass returns a CSS class suffix for a status.
func statusClass(s model.Status) string {
	switch s {
	case model.StatusOpen:
		return "open"
	case model.StatusInProgress:
		return "progress"
	case model.StatusBlocked:
		return "blocked"
	case model.StatusDeferred:
		return "deferred"
	case model.StatusClosed:
		return "closed"
	default:
		return "default"
	}
}

// typeClass returns a CSS class suffix for an issue type.
func typeClass(t model.IssueType) string {
	switch t {
	case model.TypeBug:
		return "bug"
	case model.TypeFeature:
		return "feature"
	case model.TypeEpic:
		return "epic"
	default:
		return "default"
	}
}

// truncate shortens a string to n characters.
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

var templateFuncs = template.FuncMap{
	"priorityLabel": priorityLabel,
	"statusClass":   statusClass,
	"typeClass":     typeClass,
	"truncate":      truncate,
	"upper":         strings.ToUpper,
	"replace":       strings.ReplaceAll,
	"string":        func(v any) string { return fmt.Sprintf("%s", v) },
}
