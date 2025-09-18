package gui

type Zone interface {
	// Returns the dimensions that this Widget would like available to render
	// itself. A Widget should only update the value it returns from this method
	// when its Think() method is called.
	Requested() Dims

	// Returns ex,ey, where ex and ey indicate whether this Widget is capable of
	// expanding along the X and Y axes, respectively.
	Expandable() (bool, bool)

	// Returns the region that this Widget used to render itself the last time it
	// was rendered. Should be completely contained within the region that was
	// passed to it on its last call to Draw.
	Rendered() Region
}

type BasicZone struct {
	Request_dims  Dims
	Render_region Region
	Ex, Ey        bool
}

func (bz BasicZone) Requested() Dims {
	return bz.Request_dims
}
func (bz BasicZone) Rendered() Region {
	return bz.Render_region
}
func (bz BasicZone) Expandable() (bool, bool) {
	return bz.Ex, bz.Ey
}

type CollapsableZone struct {
	Collapsed     bool
	Request_dims  Dims
	Render_region Region
	Ex, Ey        bool
}

func (cz CollapsableZone) Requested() Dims {
	if cz.Collapsed {
		return Dims{}
	}
	return cz.Request_dims
}

func (cz CollapsableZone) Rendered() Region {
	if cz.Collapsed {
		return Region{Point: cz.Render_region.Point}
	}
	return cz.Render_region
}

func (cz *CollapsableZone) Expandable() (bool, bool) {
	if cz.Collapsed {
		return false, false
	}
	return cz.Ex, cz.Ey
}
