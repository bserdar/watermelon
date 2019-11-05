package client

// Selectors are used to select a subset of hosts from an inventory

type Selector interface {
	isSelector()
}

// HasAllLabels is a selector that selects hosts containing all the labels
type HasAllLabels struct {
	Labels []string
}

// HasAnyLabel is a selector that selects hosts containing any of the labels
type HasAnyLabel struct {
	Labels []string
}

// HasNoneLabels is a selector that selects hosts containing none of the labels
type HasNoneLabels struct {
	Labels []string
}

// SelectByID is a selector that selects hosts by ID
type SelectByID struct {
	IDs []string
}

// SelectByName is a selector that selects hosts by name
type SelectByName struct {
	Names []string
}

// KeyAndValues contains a property key, and a set of values one of which should match
type KeyAndValues struct {
	Key    string
	Values []string
}

// SelectByAnyProperty is a selector that selects hosts having any of the given properties
type SelectByAnyProperty struct {
	Any []KeyAndValues
}

// SelectByAllProperty is a selector that selects hosts having all the given properties
type SelectByAllProperty struct {
	All []KeyAndValues
}

// Has returns a selector that selects hosts containing label
func Has(label string) HasAllLabels {
	return HasAllOf(label)
}

// HasAllOf returns a selector that selects hosts containing all of the labels
func HasAllOf(label ...string) HasAllLabels {
	return HasAllLabels{Labels: label}
}

// HasAnyOf returns a selector that selects hosts containing any of the labels
func HasAnyOf(label ...string) HasAnyLabel {
	return HasAnyLabel{Labels: label}
}

// HasNoneOf returns a selector that selects hosts containing none of the labels
func HasNoneOf(label ...string) HasNoneLabels {
	return HasNoneLabels{Labels: label}
}

// WithIds returns a selector that selects hosts by ID
func WithIds(id ...string) SelectByID {
	return SelectByID{IDs: id}
}

// WithNames returns a selector that selects hosts by name
func WithNames(name ...string) SelectByName {
	return SelectByName{Names: name}
}

// WithAnyProperty returns a selector that selects hosts containing any of the given properties
func WithAnyProperty(kv ...KeyAndValues) SelectByAnyProperty {
	return SelectByAnyProperty{Any: kv}
}

// WithAllProperty returns a selector that selects hosts containing all given properties
func WithAllProperty(kv ...KeyAndValues) SelectByAllProperty {
	return SelectByAllProperty{All: kv}
}

func (h HasAllLabels) isSelector()        {}
func (h HasAnyLabel) isSelector()         {}
func (h HasNoneLabels) isSelector()       {}
func (h SelectByID) isSelector()          {}
func (h SelectByName) isSelector()        {}
func (h SelectByAnyProperty) isSelector() {}
func (h SelectByAllProperty) isSelector() {}
