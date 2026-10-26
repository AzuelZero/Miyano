# Clones the exercise-dataset media (GIFs/images, ~300MB once) into backend/media.
# Idempotent: skips when the clone already exists. Media is gitignored (NOTICE.md).
param([string]$Target = "$PSScriptRoot\..\backend\media")

if (Test-Path "$Target\videos") {
    Write-Host "Media already present at $Target"
    exit 0
}

Write-Host "Cloning exercise dataset media into $Target (one-time, ~300MB)..."
git clone --depth 1 https://github.com/hasaneyldrm/exercises-dataset $Target
