# ltm/ — Local Long-Term Memory

Project-local memory managed by ltm-power.

## Commit policy: repo-portable tooling, local-private memory

**Commit:** `ltm/bin/ltm.py`, `ltm/config.json`, `ltm/manifest.json`, this README.
**Do NOT commit:** `ltm/store/`, `ltm/runtime/`, `ltm/reports/`, `ltm/snapshots/`.

If the hook uses an absolute path, review `.kiro/hooks/ltm-postturn-capture.json` before committing.

## Commands

- `python3 ltm/bin/ltm.py files --limit 10`
- `python3 ltm/bin/ltm.py health`
- `python3 ltm/bin/ltm.py checkpoint --summary "..."`
- `python3 ltm/bin/ltm.py validate`
- `python3 ltm/bin/ltm.py repair`
- `python3 ltm/bin/ltm.py purge-last --confirm`
- `python3 ltm/bin/ltm.py purge-all --confirm`
- `python3 ltm/bin/ltm.py teardown --confirm`
