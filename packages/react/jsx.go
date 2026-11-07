package react

import (
	"fmt"
)

// jsx creates a JSX host element. This is the Go equivalent of the
// React JSX runtime `jsx` function. The `type` argument is either a
// string for host elements ("box", "text", ...) or a ComponentFunc for
// user-defined components. Any trailing children are appended to the
// element's Children slice.
func jsx(type_ interface{}, props JSXProps, children ...JSXElement) JSXElement {
	return CreateJSXElement(fmt.Sprintf("%v", type_), props, children...)
}

// jsxs is an alias for jsx, matching the React JSX runtime convention
// where jsxs is used for elements with children.
var jsxs = jsx

// jsxDEV creates a JSX element in development mode. In this Go port it is
// equivalent to jsx; it is provided for API compatibility with
// react/jsx-dev-runtime.
func jsxDEV(type_ interface{}, props JSXProps) JSXElement {
	return CreateJSXElement(fmt.Sprintf("%v", type_), props)
}

// Fragment is the special element type used for Fragment elements. A Fragment
// renders its children without producing a host renderable of its own.
var FragmentJSX = Fragment

// jsxComponent creates a JSX element rendered by a function component.
func jsxComponent(fn ComponentFunc, props JSXProps, children ...JSXElement) JSXElement {
	return CreateJSXComponent(fn, props, children...)
}

// jsxFragment creates a Fragment element that wraps children without
// producing a host renderable.
func jsxFragment(children ...JSXElement) JSXElement {
	return CreateJSXFragment(children...)
}
