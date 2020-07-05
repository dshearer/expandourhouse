class WindowWidthWatcher {
    constructor(breakpoints) {
        this.breakpoints = breakpoints;
        this._listeners = [];
        this._callback = this._callback.bind(this);

        this._watchers = [];
        for (const bp of breakpoints) {
            const w = window.matchMedia(`(max-width: ${bp}px)`);
            w.breakpoint = bp;
            w.addListener(this._callback);
            this._watchers.push(w);
        }
    }

    cleanup() {
        this._watchers.forEach(w => w.removeListener(this._callback));
    }

    _callback() {
        this._listeners.forEach(L => L(this._maxWidth()));
    }

    _maxWidth() {
        var maxWidth = Number.MAX_SAFE_INTEGER;
        for (const w of this._watchers) {
            if (w.matches) {
                maxWidth = Math.min(maxWidth, w.breakpoint);
            }
        }
        return maxWidth;
    }

    addListener(L) {
        L(this._maxWidth());
        this._listeners.push(L);
    }
}

export {WindowWidthWatcher};
