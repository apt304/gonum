// Copyright ©2017 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fd

import (
	"math"
)

// Derivative estimates the derivative of the function f at the given location.
// The finite difference formula, the step size, and other options are
// specified by settings. If settings is nil, the first derivative will be
// estimated using the Forward formula and a default step size.
func Derivative(f func(float64) float64, x float64, settings *Settings) float64 {
	// Default settings.
	formula := Forward
	step := formula.Step
	var originValue float64
	var originKnown bool

	// Use user settings if provided.
	if settings != nil {
		if !settings.Formula.isZero() {
			formula = settings.Formula
			step = formula.Step
			checkFormula(formula)
		}
		if settings.Step != 0 {
			step = settings.Step
		}
		originKnown = settings.OriginKnown
		originValue = settings.OriginValue
	}

	var deriv float64
	for _, pt := range formula.Stencil {
		if originKnown && pt.Loc == 0 {
			deriv += pt.Coeff * originValue
			continue
		}
		deriv += pt.Coeff * f(x+step*pt.Loc)
	}

	return deriv / math.Pow(step, float64(formula.Derivative))
}
