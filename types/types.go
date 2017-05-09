package types

import "encoding/json"
import (
	fflow "github.com/TIBCOSoftware/flogo-lib/flow/flowdef"
)

type Microgateway struct {
	Gateway Gateway `json:"gateway"`
}

type Gateway struct {
	Name           string         `json:"name"`
	Version        string         `json:"version"`
	Description    string         `json:"description,omitempty"`
	Configurations []Config       `json:"configurations"`
	Triggers       []Trigger      `json:"triggers"`
	EventHandlers  []EventHandler `json:"event_handlers"`
	EventLinks     []EventLink    `json:"event_links"`
}

type Config struct {
	Name        string          `json:"name"`
	Type        string          `json:"type"`
	Description string          `json:"description,omitempty"`
	Settings    json.RawMessage `json:"settings"`
}

type Trigger struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type"`
	Settings    json.RawMessage `json:"settings"`
}

type EventHandler struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Reference   string          `json:"reference,omitempty"`
	Params      json.RawMessage `json:"params,omitempty"`
	Actions     []Action        `json:"actions"`
	Links       []Link          `json:"links,omitempty"`
}

type Action struct {
	ID          int             `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Type        string          `json:"type"`
	Inputs      json.RawMessage `json:"inputs,omitempty"`
	Outputs     json.RawMessage `json:"outputs,omitempty"`
}

type Link struct {
	From string `json:"from"`
	To   string `json:"to"`
	If   string `json:"if,omitempty"`
}

type EventLink struct {
	Trigger      string `json:"trigger"`
	SuccessPaths []Path `json:"success_paths"`
	ErrorPaths   []Path `json:"error_paths,omitempty"`
}

type Path struct {
	If      string `json:"if,omitempty"`
	Handler string `json:"handler"`
}

type FlowData struct {
	Flow Flow `json:"flow"`
}

type Flow struct {
	Type int `json:"type"`
	Attributes []string `json:"attributes,omitempty"`
	RootTask *fflow.TaskRep `json:"rootTask"`
}