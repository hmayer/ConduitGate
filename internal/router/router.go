package router

import (
	"conduitgate/internal/config"
	"errors"
)

type Router struct {
	rules []config.Rule
}

func NewRouter(cfg *config.Config) *Router {
	return &Router{
		rules: cfg.Rules,
	}
}

func (r *Router) Match(path string) ([]config.Destination, error) {
	for _, rule := range r.rules {
		if rule.SourcePath == path {
			return rule.Destinations, nil
		}
	}
	return nil, errors.New("no route found for path: " + path)
}
