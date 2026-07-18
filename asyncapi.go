package express

import "sort"

// ChannelDoc documents a single event/message channel (for example a Socket.IO
// event, a WebSocket topic or a message-queue subject) for inclusion in the
// generated AsyncAPI document.
type ChannelDoc struct {
	// Description explains the channel's purpose.
	Description string
	// Subscribe describes the message the server SENDS to subscribers on this
	// channel (what a client receives).
	Subscribe *MessageDoc
	// Publish describes the message clients SEND to the server on this channel
	// (what the server receives).
	Publish *MessageDoc
}

// MessageDoc documents the payload of a channel message.
type MessageDoc struct {
	// Name is a machine-friendly message name.
	Name string
	// Title is a human-friendly message title.
	Title string
	// Summary is a short explanation of the message.
	Summary string
	// ContentType is the payload media type; defaults to "application/json".
	ContentType string
	// Payload is a JSON-Schema object describing the message body.
	Payload map[string]any
	// Example is an optional example payload.
	Example any
}

// AsyncAPIDoc is the root of a generated AsyncAPI 2.6 document describing the
// application's event channels. It marshals to canonical AsyncAPI JSON.
type AsyncAPIDoc struct {
	AsyncAPI string                     `json:"asyncapi"`
	Info     OpenAPIInfo                `json:"info"`
	Servers  map[string]AsyncAPIServer  `json:"servers,omitempty"`
	Channels map[string]AsyncAPIChannel `json:"channels"`
}

// AsyncAPIServer is a single entry of the AsyncAPI servers block.
type AsyncAPIServer struct {
	URL      string `json:"url"`
	Protocol string `json:"protocol"`
}

// AsyncAPIChannel documents one channel with optional publish/subscribe
// operations.
type AsyncAPIChannel struct {
	Description string             `json:"description,omitempty"`
	Subscribe   *AsyncAPIOperation `json:"subscribe,omitempty"`
	Publish     *AsyncAPIOperation `json:"publish,omitempty"`
}

// AsyncAPIOperation is a publish or subscribe operation carrying one message.
type AsyncAPIOperation struct {
	Message AsyncAPIMessage `json:"message"`
}

// AsyncAPIMessage documents a channel message payload.
type AsyncAPIMessage struct {
	Name        string         `json:"name,omitempty"`
	Title       string         `json:"title,omitempty"`
	Summary     string         `json:"summary,omitempty"`
	ContentType string         `json:"contentType,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
	Examples    []any          `json:"examples,omitempty"`
}

// AsyncAPI builds an AsyncAPI 2.6 document from the channels registered with
// [Application.Channel]. The result can be served directly with res.JSON. When
// no channels are registered the document still validates, with an empty
// channels object.
func (app *Application) AsyncAPI() *AsyncAPIDoc {
	reg := app.docsReg()
	opts := reg.opts
	if opts.Title == "" {
		opts.withDefaults(app)
	}

	doc := &AsyncAPIDoc{
		AsyncAPI: "2.6.0",
		Info: OpenAPIInfo{
			Title:       opts.Title,
			Version:     opts.Version,
			Description: opts.Description,
		},
		Channels: map[string]AsyncAPIChannel{},
	}

	for name, cd := range reg.channels {
		ch := AsyncAPIChannel{Description: cd.Description}
		if cd.Subscribe != nil {
			ch.Subscribe = &AsyncAPIOperation{Message: toAsyncMessage(cd.Subscribe)}
		}
		if cd.Publish != nil {
			ch.Publish = &AsyncAPIOperation{Message: toAsyncMessage(cd.Publish)}
		}
		doc.Channels[name] = ch
	}
	return doc
}

// ChannelNames returns the registered channel names in sorted order.
func (app *Application) ChannelNames() []string {
	reg := app.docsReg()
	names := make([]string, 0, len(reg.channels))
	for n := range reg.channels {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func toAsyncMessage(m *MessageDoc) AsyncAPIMessage {
	ct := m.ContentType
	if ct == "" {
		ct = "application/json"
	}
	msg := AsyncAPIMessage{
		Name:        m.Name,
		Title:       m.Title,
		Summary:     m.Summary,
		ContentType: ct,
		Payload:     m.Payload,
	}
	if m.Example != nil {
		msg.Examples = []any{m.Example}
	}
	return msg
}
