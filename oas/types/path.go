package types

import "encoding/json"

// Path represents a path item.
type Path struct {
	ref         *string
	summary     *string
	description *string
	get         *Operation
	post        *Operation
	put         *Operation
	delete      *Operation
	options     *Operation
	head        *Operation
	patch       *Operation
	trace       *Operation
	servers     []*Server
	parameters  []*Parameter
}

func (p *Path) Ref(value string) *Path {
	p.ref = &value
	return p
}

func (p *Path) Summary(value string) *Path {
	p.summary = &value
	return p
}

func (p *Path) Description(value string) *Path {
	p.description = &value
	return p
}

func (p *Path) Get(build func(o *Operation)) *Path {
	if p.get == nil {
		p.get = &Operation{}
	}
	if build != nil {
		build(p.get)
	}
	return p
}

func (p *Path) Post(build func(o *Operation)) *Path {
	if p.post == nil {
		p.post = &Operation{}
	}
	if build != nil {
		build(p.post)
	}
	return p
}

func (p *Path) Put(build func(o *Operation)) *Path {
	if p.put == nil {
		p.put = &Operation{}
	}
	if build != nil {
		build(p.put)
	}
	return p
}

func (p *Path) Delete(build func(o *Operation)) *Path {
	if p.delete == nil {
		p.delete = &Operation{}
	}
	if build != nil {
		build(p.delete)
	}
	return p
}

func (p *Path) Options(build func(o *Operation)) *Path {
	if p.options == nil {
		p.options = &Operation{}
	}
	if build != nil {
		build(p.options)
	}
	return p
}

func (p *Path) Head(build func(o *Operation)) *Path {
	if p.head == nil {
		p.head = &Operation{}
	}
	if build != nil {
		build(p.head)
	}
	return p
}

func (p *Path) Patch(build func(o *Operation)) *Path {
	if p.patch == nil {
		p.patch = &Operation{}
	}
	if build != nil {
		build(p.patch)
	}
	return p
}

func (p *Path) Trace(build func(o *Operation)) *Path {
	if p.trace == nil {
		p.trace = &Operation{}
	}
	if build != nil {
		build(p.trace)
	}
	return p
}

func (p *Path) Server(url string, optionalBuild ...func(s *Server)) *Path {
	if p.servers == nil {
		p.servers = make([]*Server, 0)
	}
	s := &Server{}
	s.URL(url)
	if len(optionalBuild) > 0 && optionalBuild[0] != nil {
		optionalBuild[0](s)
	}
	p.servers = append(p.servers, s)
	return p
}

func (p *Path) Parameter(build func(param *Parameter)) *Path {
	if p.parameters == nil {
		p.parameters = make([]*Parameter, 0)
	}
	param := &Parameter{}
	if build != nil {
		build(param)
	}
	p.parameters = append(p.parameters, param)
	return p
}

func (p Path) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Ref         *string      `json:"$ref,omitempty"`
		Summary     *string      `json:"summary,omitempty"`
		Description *string      `json:"description,omitempty"`
		Get         *Operation   `json:"get,omitempty"`
		Post        *Operation   `json:"post,omitempty"`
		Put         *Operation   `json:"put,omitempty"`
		Delete      *Operation   `json:"delete,omitempty"`
		Options     *Operation   `json:"options,omitempty"`
		Head        *Operation   `json:"head,omitempty"`
		Patch       *Operation   `json:"patch,omitempty"`
		Trace       *Operation   `json:"trace,omitempty"`
		Servers     []*Server    `json:"servers,omitempty"`
		Parameters  []*Parameter `json:"parameters,omitempty"`
	}{
		Ref:         p.ref,
		Summary:     p.summary,
		Description: p.description,
		Get:         p.get,
		Post:        p.post,
		Put:         p.put,
		Delete:      p.delete,
		Options:     p.options,
		Head:        p.head,
		Patch:       p.patch,
		Trace:       p.trace,
		Servers:     p.servers,
		Parameters:  p.parameters,
	})
}

// UnmarshalJSON unmarshals the Path from JSON.
func (p *Path) UnmarshalJSON(data []byte) error {
	type Alias Path
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	return json.Unmarshal(data, &aux)
}
