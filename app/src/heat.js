const HOTTEST_HUE = 0; // red
const COOLEST_HUE = 225; // blue
const MIN_TURNOUT = 203;
const MAX_TURNOUT = 444230;

/*
    Let c = COOLEST_HUE, h = HOTTEST_HUE, t = MIN_TURNOUT, T = MAX_TURNOUT, x = turnout

    c = mt + b
    h = mT + b

    c - h = mt + b - mT - b
          = m(t - T)
    m = (c - h)/(t - T)

    c = t(c - h)/(t - T) + b
    b = c - t(c - h)/(t - T)

    H = x(c - h)/(t - T) + c - t(c - h)/(t - T)
      = (x - t)(c - h)/(t - T) + c
*/

/*
H ranges from COOLEST_HUE to HOTTEST_HUE.

Find equation mapping turnout to hue:
    H = m * turnout + b

We want:
    H = COOLEST_HUE = m * MIN_TURNOUT + b
    H = HOTTEST_HUE = m * MAX_TURNOUT + b

Solve for m and B:
    Let c = COOLEST_HUE, h = HOTTEST_HUE, t = MIN_TURNOUT, T = MAX_TURNOUT, x = turnout
    
    c = mt + b
    h = mT + b

    c - h = mt + b - mT - b
          = m(t - T)
    m = (c - h)/(t - T)

    c = t(c - h)/(t - T) + b
    b = c - t(c - h)/(t - T)

    H = x(c - h)/(t - T) + c - t(c - h)/(t - T)
      = (x - t)(c - h)/(t - T) + c

    Let K = (c - h)/(t - T).
    Let L = c - tK.
    H = (x - t)K + c = xK - tK + c = xK + L

Therefore:

    K = (COOLEST_HUE - HOTTEST_HUE)/(MIN_TURNOUT - MAX_TURNOUT)
    L = COOLEST_HUE - MIN_TURNOUT * K
    H = turnout * K + L
*/

const K = (COOLEST_HUE - HOTTEST_HUE)/(MIN_TURNOUT - MAX_TURNOUT);
const L = COOLEST_HUE - MIN_TURNOUT * K;

function hue(turnout) {
    return turnout * K + L;
}

console.assert(hue(MIN_TURNOUT) === COOLEST_HUE);
console.assert(hue(MAX_TURNOUT) === HOTTEST_HUE);

const STYLE_FORMULA = ["+", ["*", ["get", "turnout"], K], L];

export {COOLEST_HUE, HOTTEST_HUE, MIN_TURNOUT, MAX_TURNOUT, hue, STYLE_FORMULA};
