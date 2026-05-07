package jsonpath

import (
	"fmt"
	"strconv"
	"strings"
)

// segmentV3 is the new segment interface using Node/NodeList
type segmentV3 interface {
	evaluate(node Node) (NodeList, error)
	String() string
}

// wildcardSegmentV3 implements wildcard (*) for the v3 interface
type wildcardSegmentV3 struct{}

func (s *wildcardSegmentV3) evaluate(node Node) (NodeList, error) {
	switch v := node.Value.(type) {
	case []interface{}:
		result := make(NodeList, len(v))
		for i, item := range v {
			result[i] = Node{
				Location: node.Location + "[" + strconv.Itoa(i) + "]",
				Value:    item,
			}
		}
		return result, nil
	case map[string]interface{}:
		result := make(NodeList, 0, len(v))
		for key, val := range v {
			result = append(result, Node{
				Location: node.Location + "['" + escapeNormalizedPathKey(key) + "']",
				Value:    val,
			})
		}
		return result, nil
	default:
		return NodeList{}, nil
	}
}

func (s *wildcardSegmentV3) String() string { return "*" }

// nameSegmentV3 implements member access (.name) for the v3 interface
type nameSegmentV3 struct {
	name string
}

func (s *nameSegmentV3) evaluate(node Node) (NodeList, error) {
	if strings.Contains(s.name, "(") {
		return s.evaluateFunction(node)
	}
	obj, ok := node.Value.(map[string]interface{})
	if !ok {
		return NodeList{}, nil
	}
	val, exists := obj[s.name]
	if !exists {
		return NodeList{}, nil
	}
	return NodeList{{
		Location: node.Location + "['" + escapeNormalizedPathKey(s.name) + "']",
		Value:    val,
	}}, nil
}

func (s *nameSegmentV3) evaluateFunction(node Node) (NodeList, error) {
	openParen := strings.Index(s.name, "(")
	closeParen := strings.LastIndex(s.name, ")")
	if openParen == -1 || closeParen == -1 || openParen > closeParen {
		return nil, fmt.Errorf("invalid function call syntax: malformed function call")
	}
	funcName := s.name[:openParen]
	argsStr := s.name[openParen+1 : closeParen]
	fn, err := GetFunction(funcName)
	if err != nil {
		return nil, fmt.Errorf("unknown function: %s", funcName)
	}
	var args []interface{}
	if argsStr != "" {
		parsedArgs, err := parseFunctionArgs(argsStr)
		if err != nil {
			return nil, fmt.Errorf("invalid argument: %v", err)
		}
		args = append([]interface{}{node.Value}, parsedArgs...)
	} else {
		args = []interface{}{node.Value}
	}
	result, err := fn.Call(args)
	if err != nil {
		return nil, fmt.Errorf("invalid argument: %v", err)
	}
	switch v := result.(type) {
	case int:
		result = float64(v)
	case int64:
		result = float64(v)
	case int32:
		result = float64(v)
	case float32:
		result = float64(v)
	}
	return NodeList{{Location: node.Location, Value: result}}, nil
}

func (s *nameSegmentV3) String() string { return s.name }

// indexSegmentV3 implements array index access ([i]) for the v3 interface
type indexSegmentV3 struct {
	index int
}

func (s *indexSegmentV3) evaluate(node Node) (NodeList, error) {
	arr, ok := node.Value.([]interface{})
	if !ok {
		return NodeList{}, nil
	}
	idx := s.normalizeIndex(len(arr))
	if idx < 0 || idx >= len(arr) {
		return NodeList{}, nil
	}
	return NodeList{{
		Location: node.Location + "[" + strconv.Itoa(idx) + "]",
		Value:    arr[idx],
	}}, nil
}

func (s *indexSegmentV3) normalizeIndex(length int) int {
	if s.index < 0 {
		return length + s.index
	}
	return s.index
}

func (s *indexSegmentV3) String() string {
	return fmt.Sprintf("[%d]", s.index)
}

// sliceSegmentV3 implements array slice ([start:end:step]) for the v3 interface
type sliceSegmentV3 struct {
	start, end, step int
}

func (s *sliceSegmentV3) evaluate(node Node) (NodeList, error) {
	arr, ok := node.Value.([]interface{})
	if !ok {
		return NodeList{}, nil
	}
	start, end, step := s.normalizeRange(len(arr))
	indices := generateIndices(start, end, step)
	if len(indices) == 0 {
		return nil, nil
	}
	result := make(NodeList, 0, len(indices))
	for _, idx := range indices {
		if idx >= 0 && idx < len(arr) {
			result = append(result, Node{
				Location: node.Location + "[" + strconv.Itoa(idx) + "]",
				Value:    arr[idx],
			})
		}
	}
	return result, nil
}

func (s *sliceSegmentV3) normalizeRange(length int) (start, end, step int) {
	step = s.step
	if step == 0 {
		step = 1
	}
	start = s.start
	if start == 0 {
		if step > 0 {
			start = 0
		} else {
			start = length - 1
		}
	} else if start < 0 {
		start = length + start
		if start < 0 {
			if step > 0 {
				start = 0
			} else {
				start = -1
			}
		}
	} else if start >= length {
		if step > 0 {
			start = length
		} else {
			start = length - 1
		}
	}
	end = s.end
	if end == 0 {
		if step > 0 {
			end = length
		} else {
			end = -1
		}
	} else if end < 0 {
		end = length + end
		if end < 0 {
			if step > 0 {
				end = 0
			} else {
				end = -1
			}
		}
	} else if end >= length {
		if step > 0 {
			end = length
		} else {
			end = length - 1
		}
	}
	return start, end, step
}

func (s *sliceSegmentV3) String() string {
	var result strings.Builder
	result.WriteString("[")
	result.WriteString(strconv.Itoa(s.start))
	result.WriteString(":")
	result.WriteString(strconv.Itoa(s.end))
	if s.step != 1 {
		result.WriteString(":")
		result.WriteString(strconv.Itoa(s.step))
	}
	result.WriteString("]")
	return result.String()
}

// multiIndexSegmentV3 implements multi-index ([i,j,k]) for the v3 interface
type multiIndexSegmentV3 struct {
	indices []int
}

func (s *multiIndexSegmentV3) evaluate(node Node) (NodeList, error) {
	arr, ok := node.Value.([]interface{})
	if !ok {
		return NodeList{}, nil
	}
	var result NodeList
	length := len(arr)
	for _, idx := range s.indices {
		if idx < 0 {
			idx = length + idx
		}
		if idx >= 0 && idx < length {
			result = append(result, Node{
				Location: node.Location + "[" + strconv.Itoa(idx) + "]",
				Value:    arr[idx],
			})
		}
	}
	return result, nil
}

func (s *multiIndexSegmentV3) String() string {
	indices := make([]string, len(s.indices))
	for i, idx := range s.indices {
		indices[i] = strconv.Itoa(idx)
	}
	return fmt.Sprintf("[%s]", strings.Join(indices, ","))
}

// multiNameSegmentV3 implements multi-name (['a','b']) for the v3 interface
type multiNameSegmentV3 struct {
	names []string
}

func (s *multiNameSegmentV3) evaluate(node Node) (NodeList, error) {
	obj, ok := node.Value.(map[string]interface{})
	if !ok {
		return NodeList{}, nil
	}
	var result NodeList
	for _, name := range s.names {
		if val, exists := obj[name]; exists {
			result = append(result, Node{
				Location: node.Location + "['" + escapeNormalizedPathKey(name) + "']",
				Value:    val,
			})
		}
	}
	return result, nil
}

func (s *multiNameSegmentV3) String() string {
	names := make([]string, len(s.names))
	for i, name := range s.names {
		names[i] = fmt.Sprintf("'%s'", name)
	}
	return fmt.Sprintf("[%s]", strings.Join(names, ","))
}

// recursiveSegmentV3 implements recursive descent (..) for the v3 interface
type recursiveSegmentV3 struct{}

func (s *recursiveSegmentV3) evaluate(node Node) (NodeList, error) {
	var result NodeList
	result = append(result, node)
	s.collect(node, &result)
	return result, nil
}

func (s *recursiveSegmentV3) collect(node Node, result *NodeList) {
	switch v := node.Value.(type) {
	case []interface{}:
		for i, item := range v {
			child := Node{
				Location: node.Location + "[" + strconv.Itoa(i) + "]",
				Value:    item,
			}
			*result = append(*result, child)
			s.collect(child, result)
		}
	case map[string]interface{}:
		for key, val := range v {
			child := Node{
				Location: node.Location + "['" + escapeNormalizedPathKey(key) + "']",
				Value:    val,
			}
			*result = append(*result, child)
			s.collect(child, result)
		}
	}
}

func (s *recursiveSegmentV3) String() string { return ".." }

// filterSegmentV3 implements filter expression ([?expr]) for the v3 interface
type filterSegmentV3 struct {
	expr exprNode
}

func (s *filterSegmentV3) evaluate(node Node) (NodeList, error) {
	if s.expr == nil {
		return nil, nil
	}
	if node.Value == nil {
		return nil, nil
	}
	if m, ok := node.Value.(map[string]interface{}); ok {
		result, err := s.expr.evaluate(m)
		if err != nil {
			return nil, err
		}
		if result {
			return NodeList{node}, nil
		}
		return nil, nil
	}
	if arr, ok := node.Value.([]interface{}); ok {
		var results NodeList
		for i, item := range arr {
			result, err := s.expr.evaluate(item)
			if err != nil {
				return nil, err
			}
			if result {
				results = append(results, Node{
					Location: node.Location + "[" + strconv.Itoa(i) + "]",
					Value:    item,
				})
			}
		}
		return results, nil
	}
	return nil, nil
}

func (s *filterSegmentV3) String() string {
	return "[?" + exprToString(s.expr) + "]"
}

// functionSegmentV3 implements function calls for the v3 interface
type functionSegmentV3 struct {
	name string
	args []interface{}
}

func (s *functionSegmentV3) evaluate(node Node) (NodeList, error) {
	fn, err := GetFunction(s.name)
	if err != nil {
		return nil, err
	}
	var args []interface{}
	if len(s.args) == 0 {
		args = []interface{}{node.Value}
	} else {
		// 解析路径参数
		args = make([]interface{}, len(s.args))
		for i, arg := range s.args {
			if pathStr, ok := arg.(string); ok && len(pathStr) > 0 && pathStr[0] == '$' {
				// 这是一个路径引用，需要解析并求值
				resolved, err := resolvePath(pathStr, node)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve path %q: %v", pathStr, err)
				}
				args[i] = resolved
			} else {
				args[i] = arg
			}
		}
	}
	result, err := fn.Call(args)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}
	switch v := result.(type) {
	case int:
		result = float64(v)
	case int32:
		result = float64(v)
	case int64:
		result = float64(v)
	case float32:
		result = float64(v)
	}
	if arr, ok := result.([]interface{}); ok {
		nl := make(NodeList, len(arr))
		for i, item := range arr {
			nl[i] = Node{Location: node.Location, Value: item}
		}
		return nl, nil
	}
	return NodeList{{Location: node.Location, Value: result}}, nil
}

// resolvePath 解析并求值 JSONPath 表达式
func resolvePath(pathStr string, currentNode Node) (interface{}, error) {
	// 解析路径
	segments, err := parse(pathStr)
	if err != nil {
		return nil, err
	}

	// 转换为 v3 segments
	v3Segments := wrapSegments(segments)

	// 使用根节点开始求值
	// 注意：这里我们需要找到根节点的值
	// 由于我们只有当前节点，我们需要从当前节点向上遍历
	// 但在这个实现中，我们假设路径是从根开始的
	// 所以我们需要从根节点开始求值

	// 对于顶层函数调用，当前节点就是根节点
	root := currentNode

	// 求值
	nodeList := NodeList{root}
	for _, seg := range v3Segments {
		var newNodeList NodeList
		for _, n := range nodeList {
			evaluated, err := seg.evaluate(n)
			if err != nil {
				return nil, err
			}
			newNodeList = append(newNodeList, evaluated...)
		}
		nodeList = newNodeList
	}

	// 返回结果
	if len(nodeList) == 0 {
		return nil, nil
	}
	if len(nodeList) == 1 {
		return nodeList[0].Value, nil
	}
	// 返回所有值作为数组
	values := make([]interface{}, len(nodeList))
	for i, n := range nodeList {
		values[i] = n.Value
	}
	return values, nil
}

func (s *functionSegmentV3) String() string {
	args := make([]string, len(s.args))
	for i, arg := range s.args {
		switch v := arg.(type) {
		case string:
			args[i] = fmt.Sprintf("'%s'", v)
		case float64:
			args[i] = strconv.FormatFloat(v, 'f', -1, 64)
		case int:
			args[i] = strconv.Itoa(v)
		case bool:
			args[i] = strconv.FormatBool(v)
		case nil:
			args[i] = "null"
		default:
			args[i] = fmt.Sprintf("%v", v)
		}
	}
	return fmt.Sprintf("%s(%s)", s.name, strings.Join(args, ","))
}

// escapeNormalizedPathKey escapes a key for use in Normalized Path
func escapeNormalizedPathKey(key string) string {
	var result strings.Builder
	for _, r := range key {
		switch {
		case r == '\'':
			result.WriteString("\\'")
		case r == '\\':
			result.WriteString("\\\\")
		case r < 0x20:
			result.WriteString(fmt.Sprintf("\\u%04x", r))
		default:
			result.WriteRune(r)
		}
	}
	return result.String()
}

// wrapSegments converts old segment types to new segmentV3 types
func wrapSegments(oldSegs []segment) []segmentV3 {
	newSegs := make([]segmentV3, len(oldSegs))
	for i, seg := range oldSegs {
		switch s := seg.(type) {
		case *wildcardSegment:
			newSegs[i] = &wildcardSegmentV3{}
		case *nameSegment:
			newSegs[i] = &nameSegmentV3{name: s.name}
		case *indexSegment:
			newSegs[i] = &indexSegmentV3{index: s.index}
		case *sliceSegment:
			newSegs[i] = &sliceSegmentV3{start: s.start, end: s.end, step: s.step}
		case *multiIndexSegment:
			newSegs[i] = &multiIndexSegmentV3{indices: s.indices}
		case *multiNameSegment:
			newSegs[i] = &multiNameSegmentV3{names: s.names}
		case *recursiveSegment:
			newSegs[i] = &recursiveSegmentV3{}
		case *filterSegment:
			newSegs[i] = &filterSegmentV3{expr: s.expr}
		case *functionSegment:
			newSegs[i] = &functionSegmentV3{name: s.name, args: s.args}
		default:
			// Fallback: wrap in adapter
			newSegs[i] = &oldSegmentAdapter{seg: s}
		}
	}
	return newSegs
}

// oldSegmentAdapter wraps an old segment to satisfy segmentV3
type oldSegmentAdapter struct {
	seg segment
}

func (a *oldSegmentAdapter) evaluate(node Node) (NodeList, error) {
	results, err := a.seg.evaluate(node.Value)
	if err != nil {
		return nil, err
	}
	nl := make(NodeList, len(results))
	for i, r := range results {
		nl[i] = Node{Location: node.Location, Value: r}
	}
	return nl, nil
}

func (a *oldSegmentAdapter) String() string {
	return a.seg.String()
}
