package web

func (s *Server) registerRoutes() {
	s.router.Get("/", homeHandler)
}
