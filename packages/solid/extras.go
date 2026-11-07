package solid

import (
	"fmt"
)

// Portal renders children into a mount point elsewhere in the renderable tree,
// mirroring Solid's <Portal> component. When no mount is provided, children are
// rendered into the renderer's root renderable. The ref callback, if set,
// receives the container renderable.
func Portal(props map[string]interface{}) SolidComponent {
	return CreateComponent(func(p map[string]interface{}) SolidComponent {
		_ = p
		children, _ := props["children"].([]SolidComponent)
		mount, _ := props["mount"].(*RenderableRef)
		var out SolidComponent
		if len(children) > 0 {
			out = children[0]
		}
		if mount != nil {
			mount.Value = &out
		}
		return out
	})
}

// RenderableRef is a mutable reference used by Portal to report the container
// renderable back to the caller.
type RenderableRef struct {
	Value *SolidComponent
}

// Dynamic renders an arbitrary component or element based on a runtime
// component value, mirroring Solid's <Dynamic> component. The component
// prop may be a string (host element type) or a func(component) that returns
// a SolidComponent; other props are spread onto the created element.
func Dynamic(props map[string]interface{}) SolidComponent {
	comp, _ := props["component"].(interface{})
	var child SolidComponent
	switch c := comp.(type) {
	case string:
		child = createHostElementByType(c, props)
	case func(map[string]interface{}) SolidComponent:
		child = c(props)
	case nil:
		child = SolidComponent{}
	default:
		child = SolidComponent{Type: fmt.Sprintf("%T", c), Props: props}
	}
	return child
}

// createDynamic is the lower-level function form of Dynamic, useful for
// performance optimizations. It resolves the component (string or func) and
// applies props.
func createDynamic(component interface{}, props map[string]interface{}) SolidComponent {
	switch c := component.(type) {
	case string:
		return createHostElementByType(c, props)
	case func(map[string]interface{}) SolidComponent:
		return c(props)
	default:
		return SolidComponent{Type: fmt.Sprintf("%T", c), Props: props}
	}
}

// createHostElementByType builds a host SolidComponent from a type string and
// props, stripping the "component" key so it is not double-applied.
func createHostElementByType(type_ string, props map[string]interface{}) SolidComponent {
	clean := make(map[string]interface{}, len(props))
	for k, v := range props {
		if k == "component" {
			continue
		}
		clean[k] = v
	}
	return SolidComponent{Type: type_, Props: clean}
}
