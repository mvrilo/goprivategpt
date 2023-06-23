package goprivategpt

type PrivateGPT struct {
	llm    *LLM
	Server *Server
}

func New(model string, threads, tokens int) (*PrivateGPT, error) {
	llm, err := NewLLM(model, threads, tokens)
	if err != nil {
		return nil, err
	}
	server, err := NewServer(llm)
	if err != nil {
		return nil, err
	}
	return &PrivateGPT{
		llm:    llm,
		server: server,
	}, nil
}

func (p *PrivateGPT) Predict(input string) error {
	return p.llm.Predict(input)
}

func (p *PrivateGPT) Response() string {
	return p.llm.Response()
}

func (p *PrivateGPT) Shutdown() error {
	return p.Server.Shutdown()
}

func (p *PrivateGPT) Start(addr string) error {
	return p.Server.Start(addr)
}
