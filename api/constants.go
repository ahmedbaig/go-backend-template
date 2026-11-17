package api

const (
	PathSession         = "/api/v1/session"
	PathSessionDevLogin = "/api/v1/session/dev-login"
	PathEvents          = "/api/v1/events"
)

const (
	ScopePostsRead  = "posts:read"
	ScopePostsWrite = "posts:write"
)

const (
	TopicPosts       = "posts"
	EventPostCreated = "posts.created"
	EventPostUpdated = "posts.updated"
	EventPostDeleted = "posts.deleted"
)

const (
	PathPosts     = "/api/v1/posts"
	PathPost      = "/api/v1/posts/{id}"
	PathAdminInfo = "/api/v1/admin/info"
)

const RoleAdmin = "admin"
