package sh

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Data wraps a generic map for convenient access and manipulation.
type Data map[string]interface{}

// ParseJSON parses JSON string into Data.
func ParseJSON(s string) (Data, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	return Data(m), nil
}

// ParseYAML parses YAML string into Data.
func ParseYAML(s string) (Data, error) {
	var m map[string]interface{}
	if err := yaml.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	return Data(m), nil
}

// ReadJSON reads and parses a JSON file.
func ReadJSON(path string) (Data, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseJSON(string(b))
}

// ReadYAML reads and parses a YAML file.
func ReadYAML(path string) (Data, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseYAML(string(b))
}

// JSON returns the data as JSON string.
func (d Data) JSON() string {
	b, _ := json.Marshal(d)
	return string(b)
}

// JSONPretty returns the data as indented JSON string.
func (d Data) JSONPretty() string {
	b, _ := json.MarshalIndent(d, "", "  ")
	return string(b)
}

// YAML returns the data as YAML string.
func (d Data) YAML() string {
	b, _ := yaml.Marshal(d)
	return string(b)
}

// Get retrieves a value by dot-separated path (e.g., "foo.bar.baz").
// Returns nil if path doesn't exist.
func (d Data) Get(path string) interface{} {
	return getPath(d, strings.Split(path, "."))
}

// GetString retrieves a string value by path.
func (d Data) GetString(path string) string {
	v := d.Get(path)
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

// GetInt retrieves an int value by path.
func (d Data) GetInt(path string) int {
	v := d.Get(path)
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

// GetFloat retrieves a float value by path.
func (d Data) GetFloat(path string) float64 {
	v := d.Get(path)
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		f, _ := strconv.ParseFloat(val, 64)
		return f
	default:
		return 0
	}
}

// GetBool retrieves a bool value by path.
func (d Data) GetBool(path string) bool {
	v := d.Get(path)
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true" || val == "1" || val == "yes"
	case int:
		return val != 0
	case float64:
		return val != 0
	default:
		return false
	}
}

// GetData retrieves a nested Data by path.
func (d Data) GetData(path string) Data {
	v := d.Get(path)
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		return Data(m)
	}
	return nil
}

// GetArray retrieves an array by path.
func (d Data) GetArray(path string) []interface{} {
	v := d.Get(path)
	if v == nil {
		return nil
	}
	if arr, ok := v.([]interface{}); ok {
		return arr
	}
	return nil
}

// GetStringArray retrieves a string array by path.
func (d Data) GetStringArray(path string) []string {
	arr := d.GetArray(path)
	if arr == nil {
		return nil
	}
	result := make([]string, len(arr))
	for i, v := range arr {
		result[i] = fmt.Sprintf("%v", v)
	}
	return result
}

// Has checks if a path exists.
func (d Data) Has(path string) bool {
	return d.Get(path) != nil
}

// Set sets a value at the given path, creating intermediate maps as needed.
func (d Data) Set(path string, value interface{}) Data {
	parts := strings.Split(path, ".")
	setPath(d, parts, value)
	return d
}

// Delete removes a value at the given path.
func (d Data) Delete(path string) Data {
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		delete(d, path)
		return d
	}
	parent := d.GetData(strings.Join(parts[:len(parts)-1], "."))
	if parent != nil {
		delete(parent, parts[len(parts)-1])
	}
	return d
}

// Keys returns all top-level keys.
func (d Data) Keys() []string {
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	return keys
}

// Merge merges another Data into this one (shallow).
func (d Data) Merge(other Data) Data {
	for k, v := range other {
		d[k] = v
	}
	return d
}

// Clone creates a shallow copy.
func (d Data) Clone() Data {
	result := make(Data, len(d))
	for k, v := range d {
		result[k] = v
	}
	return result
}

// Filter returns a new Data with only the specified keys.
func (d Data) Filter(keys ...string) Data {
	result := make(Data)
	for _, k := range keys {
		if v, ok := d[k]; ok {
			result[k] = v
		}
	}
	return result
}

// Exclude returns a new Data without the specified keys.
func (d Data) Exclude(keys ...string) Data {
	result := d.Clone()
	for _, k := range keys {
		delete(result, k)
	}
	return result
}

// Pick extracts values at the given paths into a new Data.
func (d Data) Pick(paths ...string) Data {
	result := make(Data)
	for _, path := range paths {
		if v := d.Get(path); v != nil {
			result.Set(path, v)
		}
	}
	return result
}

// Helper functions

func getPath(data interface{}, parts []string) interface{} {
	if len(parts) == 0 || data == nil {
		return data
	}

	key := parts[0]

	switch v := data.(type) {
	case map[string]interface{}:
		return getPath(v[key], parts[1:])
	case Data:
		return getPath(v[key], parts[1:])
	case []interface{}:
		// Support array indexing like "items.0.name"
		idx, err := strconv.Atoi(key)
		if err != nil || idx < 0 || idx >= len(v) {
			return nil
		}
		return getPath(v[idx], parts[1:])
	default:
		return nil
	}
}

func setPath(data map[string]interface{}, parts []string, value interface{}) {
	if len(parts) == 0 {
		return
	}

	key := parts[0]

	if len(parts) == 1 {
		data[key] = value
		return
	}

	// Create intermediate map if needed
	next, ok := data[key].(map[string]interface{})
	if !ok {
		next = make(map[string]interface{})
		data[key] = next
	}
	setPath(next, parts[1:], value)
}

// Array wraps a slice for convenient operations.
type Array []interface{}

// ParseJSONArray parses JSON string into Array.
func ParseJSONArray(s string) (Array, error) {
	var arr []interface{}
	if err := json.Unmarshal([]byte(s), &arr); err != nil {
		return nil, err
	}
	return Array(arr), nil
}

// JSON returns the array as JSON string.
func (a Array) JSON() string {
	b, _ := json.Marshal(a)
	return string(b)
}

// JSONPretty returns the array as indented JSON string.
func (a Array) JSONPretty() string {
	b, _ := json.MarshalIndent(a, "", "  ")
	return string(b)
}

// Len returns the length of the array.
func (a Array) Len() int {
	return len(a)
}

// Get returns the element at index.
func (a Array) Get(i int) interface{} {
	if i < 0 || i >= len(a) {
		return nil
	}
	return a[i]
}

// GetData returns element at index as Data.
func (a Array) GetData(i int) Data {
	v := a.Get(i)
	if v == nil {
		return nil
	}
	if m, ok := v.(map[string]interface{}); ok {
		return Data(m)
	}
	return nil
}

// Map applies a function to each element.
func (a Array) Map(fn func(interface{}) interface{}) Array {
	result := make(Array, len(a))
	for i, v := range a {
		result[i] = fn(v)
	}
	return result
}

// Filter returns elements where predicate is true.
func (a Array) Filter(fn func(interface{}) bool) Array {
	var result Array
	for _, v := range a {
		if fn(v) {
			result = append(result, v)
		}
	}
	return result
}

// Pluck extracts a field from each object element.
func (a Array) Pluck(key string) Array {
	var result Array
	for _, v := range a {
		if m, ok := v.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				result = append(result, val)
			}
		}
	}
	return result
}

// PluckStrings extracts a string field from each object element.
func (a Array) PluckStrings(key string) []string {
	var result []string
	for _, v := range a {
		if m, ok := v.(map[string]interface{}); ok {
			if val, exists := m[key]; exists {
				result = append(result, fmt.Sprintf("%v", val))
			}
		}
	}
	return result
}

// First returns the first element or nil.
func (a Array) First() interface{} {
	if len(a) == 0 {
		return nil
	}
	return a[0]
}

// Last returns the last element or nil.
func (a Array) Last() interface{} {
	if len(a) == 0 {
		return nil
	}
	return a[len(a)-1]
}
