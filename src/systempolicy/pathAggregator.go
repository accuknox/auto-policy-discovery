package systempolicy

import (
	"regexp"
	"sort"
	"strings"

	"github.com/accuknox/auto-policy-discovery/src/libs"
	types "github.com/accuknox/auto-policy-discovery/src/types"
)

var WildPathDigit string = "/[0-9]+"
var WildPathDigitLeaf string = "/[0-9^/]+"
var WildPathChar string = "/.+"
var WildPathCharLeaf string = "/.[^/]+"

var WildPaths []string

const Threshold = 3

func init() {
	WildPaths = []string{WildPathDigit, WildPathChar}
}

// ============================ //
// == PathNode and functions == //
// ============================ //

// Node Structure
type Node struct {
	path  string
	isDir bool

	depth      int
	touchCount int
	childNodes []*Node
}

// MergedNode Structure
type MergedNode struct {
	path string

	depth      int
	touchCount int
}

// HTTPDst Structure
type HTTPDst struct {
	Namespace   string
	MatchLabels string
	ToPorts     []types.SpecPort
	HTTPTree    map[string]map[string]*Node
}

func (n *Node) generatePaths(results map[string]bool, parentPath string) {
	for _, childNode := range n.childNodes {
		childNode.generatePaths(results, parentPath+n.path)
	}

	// if this is the leaf node
	if len(n.childNodes) == 0 {
		if n.isDir { // is matchDirectories
			results[parentPath+n.path] = true
		} else { // is matchPaths
			results[parentPath+n.path] = false
		}
	}
}

func (n *Node) insert(paths []string) {
	for _, path := range paths {
		child := n.findChildNode(path, n.depth+1)

		if child == nil {
			newChild := &Node{
				depth:      n.depth + 1,
				path:       path,
				touchCount: 1,
				childNodes: []*Node{},
			}

			n.childNodes = append(n.childNodes, newChild)
			newChild.insert(paths[1:])
		} else {
			child.touchCount++
			child.insert(paths[1:])
		}

		break
	}
}

func (n *Node) aggregateChildNodes() {
	// depth first search
	for _, childNode := range n.childNodes {
		childNode.aggregateChildNodes()
	}

	// #child nodes > threshold --> aggreagte it, and make matchDirectories
	if len(n.childNodes) > Threshold {
		n.childNodes = nil
		n.touchCount = 1 // reset touch count
		n.isDir = true
	}
}

func (n *Node) makeChildNodeToDir() {
	// depth first search
	for _, childNode := range n.childNodes {
		childNode.makeChildNodeToDir()
	}

	// #child nodes > threshold --> aggreagte it, and make matchDirectories
	if len(n.childNodes) == 0 {
		n.touchCount = Threshold + 1 // reset touch count
		n.isDir = true
	}
}

func (n *Node) findChildNode(path string, depth int) *Node {
	for _, child := range n.childNodes {
		// case 1: regex matching
		if libs.ContainsElement(WildPaths, child.path) && child.depth == depth {
			r, _ := regexp.Compile(child.path)
			if r.FindString(path) == path {
				return child
			}
			// case 2: exact matching
		} else if child.path == path && child.depth == depth {
			return child
		}
	}

	return nil
}

// ===================== //
// == Build Path Tree == //
// ===================== //

func buildPathTree(treeMap map[string]*Node, paths []string) {
	pattern, _ := regexp.Compile("(/.[^/]*)")

	// sorting paths
	sort.Strings(paths)

	// iterate paths
	for _, path := range paths {
		if path == "/" { // rootpath
			continue
		}

		// example: /usr/lib/python2.7/UserDict.py
		// 			--> '/usr', '/lib', '/python2.7', '/UserDict.py'
		//			in this case, '/usr' is rootNode
		tokenizedPaths := pattern.FindAllString(path, -1)
		if len(tokenizedPaths) == 0 {
			continue
		}

		rootPath := tokenizedPaths[0]

		if rootNode, ok := treeMap[rootPath]; !ok {
			newRoot := &Node{
				depth:      0,
				path:       rootPath,
				touchCount: 1,
				childNodes: []*Node{},
			}

			newRoot.insert(tokenizedPaths[1:])
			treeMap[rootPath] = newRoot
		} else {
			rootNode.touchCount++
			rootNode.insert(tokenizedPaths[1:])
		}
	}
}

func AggregatePaths(paths []string) []SysPath {
	treeMap := map[string]*Node{}

	// step 1: build path tree
	// paths := []string{"/usr/lib/python2.7/UserDict.py", "/usr/lib/python2.7/UserDict.pyo"}
	// -->
	// /usr 0 461
	// /lib 1 328
	// 		/python2.7 2 328
	// 				/UserDict.py 3 1
	// 				/UserDict.pyo 3 1
	// ...
	buildPathTree(treeMap, paths)

	// for root, childs := range treeMap {
	// 	fmt.Println(root)
	// 	printTree(childs)
	// }

	// step 2: aggregate path
	for _, root := range treeMap {
		root.aggregateChildNodes()
	}

	// for root, childs := range treeMap {
	// 	fmt.Println(root)
	// 	printTree(childs)
	// }

	// step 3: generate tree -> path string
	aggregatedPaths := map[string]bool{}
	for _, root := range treeMap {
		root.generatePaths(aggregatedPaths, "")
	}

	// step 4: make result
	results := []SysPath{}
	for path, isDir := range aggregatedPaths {
		if isDir && !strings.HasSuffix(path, "/") {
			path = path + "/"
		}
		sysPath := SysPath{
			Path:  path,
			isDir: isDir,
		}
		results = append(results, sysPath)
	}

	return results
}

// ========================================= //
// == Update Duplicated Paths/Directories == //
// ========================================= //

func MergeAndAggregatePaths(dirs []string, paths []string) []SysPath {
	treeMap := map[string]*Node{}

	// step 1: build path tree from matchDirectories
	// paths := []string{"/usr/lib/python2.7/UserDict.py", "/usr/lib/python2.7/UserDict.pyo"}
	// -->
	// /usr 0 461
	// /lib 1 328
	// 		/python2.7 2 328
	// 				/UserDict.py 3 1
	// 				/UserDict.pyo 3 1
	// ...
	buildPathTree(treeMap, dirs)
	for _, root := range treeMap {
		root.makeChildNodeToDir()
	}

	// step 2: append matchPaths to the path tree
	buildPathTree(treeMap, paths)

	// step 3: aggregate new paths/directories
	for _, root := range treeMap {
		root.aggregateChildNodes()
	}

	// step 4: generate tree -> path string
	aggregatedPaths := map[string]bool{}
	for _, root := range treeMap {
		root.generatePaths(aggregatedPaths, "")
	}

	// step 5: make result
	results := []SysPath{}
	for path, isDir := range aggregatedPaths {
		sysPath := SysPath{
			Path:  path,
			isDir: isDir,
		}
		results = append(results, sysPath)
	}

	return results
}
