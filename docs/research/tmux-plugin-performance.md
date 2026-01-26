# Performance Strategies in Popular Tmux Status Line Plugins

Research into how popular tmux plugins handle status line performance, conducted to inform the design of tmux-asdf-status.

## Plugins Reviewed

1. [tmux-cpu](https://github.com/tmux-plugins/tmux-cpu)
2. [tmux-battery](https://github.com/tmux-plugins/tmux-battery)
3. [tmux-mem-cpu-load](https://github.com/thewtex/tmux-mem-cpu-load)
4. [tmux-online-status](https://github.com/tmux-plugins/tmux-online-status)
5. [tmux-prefix-highlight](https://github.com/tmux-plugins/tmux-prefix-highlight)

---

## Detailed Findings

### tmux-cpu

**Caching:** Most sophisticated among plugins studied. Uses file-based caching in `/tmp/tmux-$EUID-cpu/`:

```bash
cached_eval() {
  local command key val
  command="$1"
  key="$(basename "$command")"
  val="$(get_cache_val "$key")"
  if [ -z "$val" ]; then
    put_cache_val "$key" "$($command "${@:2}")"
  else
    echo -n "$val"
  fi
}
```

- Cache file stores timestamp on line 1, output on subsequent lines
- 2-second default TTL
- Directory uses mode 0700 with `$EUID` for security

**Expensive Operation Handling:**
- Prefers `iostat` or `sar` over `ps aux`
- Platform-specific code paths (Linux, macOS, FreeBSD, OpenBSD, Cygwin)
- Stream processing with `sed`, `tail`, `awk` pipelines

**Update Strategy:** Relies on tmux's `status-interval`; cache prevents redundant calls within window.

---

### tmux-battery

**Caching:** None. Fresh fetch on every update.

**Expensive Operation Handling:**
- Early termination via `command_exists` checks
- Targeted output parsing: `grep -o "[0-9]\{1,3\}%"`
- Fallback chain: pmset -> acpi -> upower -> termux -> apm
- README warns against `upower` "due to ongoing CPU usage issues"

**Update Strategy:** Recommends `status-interval 15`. Documentation notes 15-60 second delay for icon changes after unplugging.

---

### tmux-mem-cpu-load

**Caching:** None (different approach as compiled C++ binary).

**Expensive Operation Handling:**
- Compiled binary inherently faster than shell scripts
- Pre-computed color lookup tables (`luts.h`)
- Direct system calls: `getloadavg()`, `/proc/stat`
- Dynamic precision adjustment

**Update Strategy:** The `--interval` argument must match tmux's `status-interval`:

```cpp
cpu_usage_delay = atoi(optarg) * 1000000 - 10000;  // microseconds
```

The program samples CPU twice with this delay between reads. Mismatched intervals cause inaccurate measurements.

---

### tmux-online-status

**Caching:** None. Fresh network check every update.

**Expensive Operation Handling:**
- Single packet: `ping -c 1`
- Configurable timeout (default: 3 seconds)
- Platform-specific ping flags

**Update Strategy:** Recommends `status-interval 5`.

**Performance Concern:** Synchronous blocking ping on every update. On unstable connections, this causes noticeable status bar lag. Most expensive plugin per-update due to network I/O.

---

### tmux-prefix-highlight

**Caching:** None needed due to architectural approach.

**Expensive Operation Handling:** Most optimized design studied:

1. Delegates logic to tmux's format string engine:
   ```
   #{?client_prefix,$prefix_mode,$fallback}
   ```
2. Configuration read once at initialization
3. No subprocess calls during rendering
4. Lazy evaluation via tmux format strings

**Update Strategy:** Event-driven, not polling-based. Zero overhead when prefix inactive.

---

## Comparison Table

| Plugin | Caching | Primary Optimization | Recommended Interval |
|--------|---------|---------------------|---------------------|
| tmux-cpu | Yes (2s file-based) | Cache expensive system calls | 2s effective minimum |
| tmux-battery | No | Tool prioritization, early exits | 15 seconds |
| tmux-mem-cpu-load | No | Compiled C++ binary | Configurable (must match tmux) |
| tmux-online-status | No | Single-packet ping, timeout | 5 seconds |
| tmux-prefix-highlight | No | Delegates to tmux format engine | N/A (event-driven) |

---

## Summary

**Most plugins do not cache.** They rely on:
- Recommending longer `status-interval` values (5-15 seconds)
- Early exits and tool prioritization
- Efficient parsing with targeted regex/awk

**tmux-cpu is the exception** with proper file-based caching. Its approach:
- Cache in `/tmp/` with user-specific directory
- Timestamp-based TTL (2 seconds)
- Cache key derived from command name

**tmux-prefix-highlight takes a different approach** by delegating all logic to tmux's format string engine, avoiding subprocess overhead entirely.

**tmux-mem-cpu-load solves performance via compilation** rather than caching, achieving speed through C++ instead of shell scripts.

---

## Conclusions for tmux-asdf-status

### Adopt from tmux-cpu
- **File-based caching pattern:** Store results in `/tmp/tmux-asdf-status-$EUID/`
- **Timestamp-based TTL:** Check cache age before regenerating
- **Security considerations:** Mode 0700, user-specific directory

### Adopt from tmux-battery
- **Early exit pattern:** Check for `.tool-versions` existence before any processing
- **Tool prioritization:** Use shell builtins over external commands where possible

### Consider from tmux-prefix-highlight
- **Minimize subprocess calls:** Where possible, batch tmux option queries rather than calling `tmux show-option` repeatedly

### Additional Recommendations

1. **Cache invalidation by mtime:** Unlike tmux-cpu which uses pure TTL, we should also invalidate when `.tool-versions` changes. This provides faster response to actual changes while still benefiting from caching.

2. **Configurable TTL:** Default 5 seconds is reasonable (longer than tmux-cpu's 2s since tool versions change less frequently than CPU usage).

3. **Cache the tool-versions path:** Once we find `.tool-versions` by walking up the directory tree, cache the path. Subsequent calls from the same directory can skip the walk.

4. **Avoid network I/O:** Unlike tmux-online-status, we have no network requirements. Keep it that way.

5. **Shell script is appropriate:** Our operation (file reads, string processing) doesn't warrant a compiled binary like tmux-mem-cpu-load. The caching strategy is sufficient.

### Proposed Cache Structure

```
/tmp/tmux-asdf-status-$EUID/
├── cache_<hash>           # Cached output for directory
│   ├── line 1: timestamp
│   ├── line 2: .tool-versions mtime
│   └── line 3+: formatted output
└── ...
```

Cache hit conditions:
1. Cache file exists
2. Age < TTL (default 5s)
3. `.tool-versions` mtime unchanged

This hybrid approach (TTL + mtime) provides both performance and responsiveness.
