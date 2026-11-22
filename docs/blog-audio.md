# Blog audio: generation and publishing

The English blog narration is generated locally with Ollama Phonon and served
from the public Google Cloud Storage bucket `go-brianbondy-blog-audio`. The
small sentence timing files live in `static/audio/blog/` and the published MP3
URLs live in `data/blogAudioManifest.json`. The site only shows a player for a
post when it has a published manifest entry, or when both files exist locally
for an unpublished post.

## 1. Install the local tools

Use an Apple Silicon Mac with Go, Python 3, FFmpeg, and Ollama 0.40.9 or newer.
On macOS with Homebrew:

```bash
brew install go ffmpeg ollama
brew services start ollama
```

`make blog-audio` checks these dependencies. If Ollama or FFmpeg is missing, it
prints the install command. If Ollama is installed but not running, it prints
how to start it. When the `phonon` model is missing, the generator downloads it
automatically with `ollama pull phonon`. The model is about 112 MB. The code
and model are Apache 2.0 licensed. `ollama show phonon` lists available voices.

## 2. Generate audio

For one post:

```bash
make blog-audio BLOG_AUDIO_ID=186
```

For every post in `data/blogPostManifest.json`:

```bash
make blog-audio
```

The generator saves each post's MP3 and timing JSON under `static/audio/blog/`.
MP3s are ignored by Git and `.gcloudignore`; timing JSON is committed with the
site. Repeating the command skips unchanged posts and reuses sentence audio in
`.cache/blog-audio/`, which is also excluded from Git and App Engine uploads.
An interrupted batch can be resumed. The current
171-post archive is 241 MB of MP3s (8.38 hours of narration). Future posts will
add to that total.

To try another voice or force new speech for one post:

```bash
python3 scripts/generate_blog_audio.py --id 186 --voice Hadley --force
```

Phonon does not accept emotion instructions or return word timestamps. The
generator times each chunk from its actual PCM length, so the player highlights
sentences rather than individual words. It narrates prose, headings, and list
items; code blocks and media are skipped.

Before publishing, listen to `static/audio/blog/186.mp3` or open a newly
generated post locally with `go run .`. If port 8080 is occupied, run
`PORT=8081 go run .`. A post that is already published uses its Cloud Storage
audio URL on the site; listen to its newly generated local MP3 directly before
publishing an update.

## 3. Publish to Cloud Storage

The dedicated public bucket already exists in project `go-brianbondy` in
`us-east1`. To publish, the Google Cloud CLI must be authenticated to an
account with permission to upload to that bucket. Check with:

```bash
gcloud auth list
gcloud config get-value project
```

If needed, run `gcloud auth login` and `gcloud config set project go-brianbondy`.
Install the [Google Cloud CLI](https://cloud.google.com/sdk/docs/install) if
`gcloud` is not available.
Publish a single generated post or every generated post:

```bash
make publish-blog-audio BLOG_AUDIO_ID=186
make publish-blog-audio
```

The publisher checks that each MP3 matches its timing JSON, uploads a
content-hashed filename with `audio/mpeg` and long-lived cache headers, verifies
that the public URL serves the file, then updates `data/blogAudioManifest.json`.
It skips posts that already have a matching manifest entry. The bucket can be
overridden with `BLOG_AUDIO_BUCKET=your-bucket` if needed. To inspect planned
uploads without writing to Cloud Storage, run:

```bash
python3 scripts/publish_blog_audio.py --bucket go-brianbondy-blog-audio --all --dry-run
```

The MP3 remains in `static/audio/blog/` locally so it can be replayed or
republished, but it is not uploaded as part of an App Engine deployment.
Previously published versioned MP3 objects are not deleted automatically.

## 4. Deploy the site

Commit the source changes, timing JSON files, and
`data/blogAudioManifest.json`. The MP3s stay out of Git. Run `make test`, then
deploy with `make deploy` when ready. App Engine serves the timing JSON as
static files; the MP3s play directly from Cloud Storage. Restart a local
`go run .` server after publishing so it reloads the URL manifest.

To update a post later, edit its Markdown, regenerate and review the MP3,
publish it, and deploy the updated JSON and manifest.
The filename changes when the MP3 changes, so browser caches do not mix old
audio with new timing data.

## One-time bucket setup

This bucket has already been created. If it ever needs to be recreated, use a
new globally unique bucket name and update `BLOG_AUDIO_BUCKET` when publishing:

```bash
gcloud storage buckets create gs://YOUR_UNIQUE_BUCKET \
  --project=go-brianbondy --location=us-east1 \
  --default-storage-class=STANDARD \
  --uniform-bucket-level-access --no-public-access-prevention
gcloud storage buckets add-iam-policy-binding gs://YOUR_UNIQUE_BUCKET \
  --member=allUsers --role=roles/storage.objectViewer
```

The bucket is dedicated to public narration files. Do not put private files
in it. Public access is required for browsers to play the MP3s without signing
in. The publisher writes only MP3 objects, and Git keeps the associated timing
files and URLs.
