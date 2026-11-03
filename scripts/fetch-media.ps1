# Clones the exercise-dataset media (GIFs/images, ~300MB once) into backend/media.
# Idempotent: skips when the clone already exists. Media is gitignored (NOTICE.md).
# Supply-chain pin (audit 03/11): the clone must land exactly on $PinnedCommit —
# bumping the dataset means bumping the pin consciously.
param(
    [string]$Target = "$PSScriptRoot\..\backend\media",
    [string]$PinnedCommit = "7455efae41b330c265e7cd4b78dfa848e7ce5ebd"
)

function Assert-Pinned([string]$Dir) {
    $head = "$(git -C $Dir rev-parse HEAD)"
    if ($head -ne $PinnedCommit) {
        throw "Media at $Dir is at $head but the pin is $PinnedCommit. Update -PinnedCommit consciously or delete the folder."
    }
}

if (Test-Path "$Target\videos") {
    Assert-Pinned $Target
    Write-Host "Media already present at $Target (pinned commit OK)"
    exit 0
}

Write-Host "Cloning exercise dataset media into $Target (one-time, ~300MB)..."
git clone --quiet --no-checkout https://github.com/hasaneyldrm/exercises-dataset $Target
if ($LASTEXITCODE -ne 0) { throw "git clone failed" }
git -C $Target fetch --quiet --depth 1 origin $PinnedCommit
if ($LASTEXITCODE -ne 0) { throw "git fetch of pinned commit $PinnedCommit failed (pin outdated?)" }
git -C $Target checkout --quiet $PinnedCommit
if ($LASTEXITCODE -ne 0) { throw "git checkout of $PinnedCommit failed" }
Assert-Pinned $Target
Write-Host "Media ready at $Target (pinned to $PinnedCommit)"
