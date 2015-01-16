package main

type SiteInfo struct {
	Title              string
	Secret             string
	PostsPerPage       int
	ThreadsPerPage     int
	TagResultsPerPage  int
	TagsPerPage        int
	MaxPostLength      int
	Footer             []string
	PostFooter         string
	FullVersionDomain  string
	LightVersionDomain string
}

type AdminConfig struct {
	EnablePath   bool
	EnableDomain bool
	Path         string
	Domain       string
	Admins       map[string]string
}

type Database interface {
	Init(string, string) Database
	Close()
	NewThread(User, string, string, []string) (string, error)
	ReplyThread(*Thread, User, string) (int, error)
	GetThreadList(string, int, int, []string) ([]Thread, error)
	GetThread(string) (Thread, error)
	GetThreadById(interface{}) (Thread, error)
	GetPost(interface{}) (Post, error) // Consider deprecating
	GetPosts(*Thread, int, int) ([]Post, error)
	PostCount(*Thread) (int, error)
	GetPopularTags(int, int, []string) ([]Tag, error)
	IncTag(string, interface{}) error // Move to internal only..?
	DecTag(string) error              // Move to internal only.. ?
	EditPost(interface{}, string) error
	DeletePost(interface{}, bool) error
	SetThreadTags(interface{}, []string) error
	GetMatchingTags(string, int, int, []string) ([]Tag, error)
}
