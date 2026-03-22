# Deep Research Report: Container Image Diff Tool

**Date:** 2026-03-09
**Optimization Target:** GitHub stars and developer reputation
**Founder Profile:** Solo dev, Go + Docker/K8s experience, side-project time

---

## 1. Executive Summary

**Verdict: BUILD (with sharp differentiation)**
**Confidence: 7/10**

There is a genuine, validated gap in the container tooling ecosystem: no single tool provides a beautiful, "git diff"-style comparison of two container images showing file-level changes, size impact, and security-relevant mutations. The demand signal is real (users have asked for this on dive's HN threads since 2018, Google's container-diff was archived in 2024, and the only active alternative -- diffoci -- has just 561 stars with a narrow reproducible-builds focus). The founder has the exact right skill set (Go, Docker, infra). However, execution must be razor-sharp: this is a crowded adjacent space, and the tool must be dramatically better at the *specific* job of two-image comparison than everything else, or it will be perceived as "just another container tool."

The primary risk is not competition -- it is apathy. Most developers don't compare images often enough to seek a dedicated tool. The path to stars requires making the tool so visually stunning and useful in CI/CD pipelines that it spreads through screenshots, blog posts, and "TIL" moments.

---

## 2. Competitive Landscape

### Direct Competitors

| Tool | Stars | Language | Status | What It Does | Key Limitation for Your Idea |
|------|-------|----------|--------|-------------|------------------------------|
| **dive** (wagoodman) | 53.5K | Go | Active (slow) | Explores layers of a *single* image | Cannot compare two images at all. Single-image only. |
| **container-diff** (Google) | 3.8K | Go | **Archived** (Mar 2024) | Compares two images (files, packages, size) | Dead. README says "try diffoci." No visual output. |
| **diffoci** | 561 | Go | Active | Semantic diff of two OCI images | Focused on reproducible builds. Output is a table of hashes, not human-friendly. No file-content diff. No security analysis. |
| **docker scout compare** | N/A (Docker CLI) | N/A | Active (experimental) | Compares CVEs and packages between two images | Vulnerabilities only. No file-level diff. No layer analysis. Requires Docker subscription for full features. |
| **dredge** | 44 | C# | Active | Registry client with image file comparison | Tiny project. Requires .NET runtime. Launches external diff tool. |
| **moul/docker-diff** | 157 | Shell | Abandoned (2016) | Compares two images via filesystem extraction | Shell script from 2016. No longer relevant. |
| **SlimToolkit (xray)** | 23K | Go | Active | Analyzes single image, reverse-engineers Dockerfile | Single-image only. Comparison is SaaS-only feature. |
| **docker diff** (built-in) | N/A | N/A | N/A | Shows changes in a running *container* vs its base image | Container only, not images. Completely different use case. |

### Key Takeaway

The market has a **dead center** right where your tool would sit:
- **dive** = single image explorer (53K stars, proving massive demand for image analysis)
- **container-diff** = two-image comparison (archived, proving Google couldn't sustain it as a side project but validating the concept)
- **diffoci** = two-image comparison (active but niche, focused on reproducible builds, not developer UX)
- **docker scout compare** = vulnerability comparison only (not file-level)

**Nobody is doing: beautiful, file-level, security-aware, two-image diff with visual terminal output.** This is the gap.

### Indirect / Tangential Tools

| Tool | Relevance |
|------|-----------|
| **skopeo** | Can inspect manifests but has no diff feature |
| **crane** (go-containerregistry) | Layer manipulation library, no diff CLI |
| **trivy / grype / syft** | Vulnerability scanning, no image-to-image file diff |
| **sbomlyze** | SBOM diff (package-level, not file-level) |

### Critical Intelligence: wagoodman & Anchore

Alex Goodman (dive's author) is Tech Lead on Anchore's OSS team and maintains dive, syft, grype, and **stereoscope** (the Go library that powers dive's layer reading). Stereoscope is the exact library you would want to use -- or compete with. Goodman committed (March 2025) to establishing a roadmap and external maintainers for dive, but has not announced any two-image comparison feature. The risk of dive adding this feature exists but seems low given Anchore's focus on security tooling (syft/grype), not developer UX tools.

---

## 3. Technical Implementation

### Architecture

```
                    +------------------+
                    |   CLI Interface  |  (cobra + bubbletea/lipgloss)
                    +--------+---------+
                             |
                    +--------v---------+
                    |   Diff Engine    |
                    +--------+---------+
                             |
              +--------------+--------------+
              |              |              |
     +--------v----+  +-----v------+  +----v--------+
     | Image Loader |  | Layer Diff |  | Report Gen  |
     | (registry,   |  | (file tree |  | (terminal,  |
     |  daemon,     |  |  compare,  |  |  JSON, MD,  |
     |  tarball)    |  |  whiteout) |  |  CI output) |
     +-------------+  +------------+  +-------------+
```

### Core Libraries

| Library | Purpose | Stars |
|---------|---------|-------|
| **google/go-containerregistry** | Pull images from registries, daemons, tarballs. Handles auth, multi-arch. | 3.8K |
| **anchore/stereoscope** | Read layer file trees, build squashed filesystem. Handles whiteout files. | ~200 |
| **charmbracelet/bubbletea** | Beautiful interactive TUI | 30K+ |
| **charmbracelet/lipgloss** | Terminal styling (colors, borders, tables) | 9K+ |
| **cobra** | CLI framework | 40K+ |

**Build vs. Buy Decision:** Use `go-containerregistry` for image access (it is the standard). For layer tree diffing, you could use `stereoscope` OR build your own (stereoscope is heavier than you need and ties you to Anchore's ecosystem). For a focused diff tool, building a lightweight layer-to-filetree converter on top of `go-containerregistry`'s raw layer tar readers may be better for control and performance.

### OCI Layer Diffing: How It Works

1. **Pull manifests** for both images (cheap -- just JSON, a few KB)
2. **Compare layer digests** -- identical layers can be skipped entirely (huge optimization)
3. For differing layers, **stream each layer tar** and build an in-memory file tree (path, size, permissions, hash)
4. Handle **whiteout files** (`.wh.` prefix = deletion, `.wh..wh..opq` = opaque directory wipe)
5. Build **squashed file trees** for each image (cumulative state after all layers)
6. **Diff the squashed trees**: added files, removed files, modified files (size change, permission change, content hash change)
7. Optionally **diff individual layers** for layer-by-layer analysis

### Multi-Arch Handling

OCI multi-arch images use an **Image Index** (manifest list) pointing to per-platform manifests. The tool should:
- Default to the host platform (like Docker does)
- Support `--platform linux/amd64` flag to select specific platform
- Show manifest index summary when platforms differ between images

### Performance Considerations

- **Manifest-only mode** (`--quick`): Compare layer digests only, no content download. Instant for registry images.
- **Streaming comparison**: Don't download entire layers to disk. Stream tar entries and build file trees in memory.
- **Shared layer skip**: If images share base layers (same digest), skip them entirely.
- For multi-GB images: expect 30-60s for full file-level diff. This is acceptable -- container-diff had similar performance.

### MVP Scope (4-6 weeks)

**Week 1-2: Core engine**
- Pull two images (local daemon + registry)
- Build squashed file trees
- Compute file-level diff (added/removed/modified)
- Handle whiteout files

**Week 3-4: Output & UX**
- Beautiful terminal output with lipgloss (color-coded: green=added, red=removed, yellow=modified)
- Size impact summary (total size change, per-layer breakdown)
- Security highlights (new SUID binaries, permission changes, new executables)
- JSON output for CI/CD

**Week 5-6: Polish & Distribution**
- Multi-arch support
- `--quick` manifest-only mode
- Homebrew formula
- GitHub Action
- README with beautiful screenshots/GIF

### Post-MVP Features (ordered by star potential)

1. **Interactive TUI mode** (bubbletea) -- browse diff like dive's UI
2. **Package-level diff** (detect apt/apk/yum package changes)
3. **Docker plugin** (`docker imgdiff image1 image2`)
4. **Dockerfile attribution** (which Dockerfile instruction caused each change)
5. **SBOM-aware diff** (integrate with syft for richer package info)

---

## 4. Market & Adoption

### Total Addressable Audience

| Metric | Number | Source |
|--------|--------|--------|
| Docker Hub accounts | 7.3M | Docker (2025) |
| Docker adoption among IT pros | 92% | Docker State of App Dev 2025 |
| Professional developers using Docker | 71.1% | Stack Overflow 2025 |
| Monthly container pulls | 13B | Docker Index |
| Container security market | ~$3B (2025) | Multiple analyst firms |

### Why Dive Got 53K Stars

Dive's growth followed a pattern common to successful developer tools:
1. **Show HN post** (Nov 25, 2018): 684 points, 42 comments. Community reaction: "I always missed but never actually thought about" this tool.
2. **Visual appeal**: The TUI with color-coded file trees was immediately screenshottable.
3. **Solves a real pain**: "My Java image went up by about 4x the library size" -- developers couldn't figure out why images were bloated.
4. **CI integration**: The `CI=true` mode gave it a second life in pipelines.
5. **Repeated HN appearances**: Got 464 points again in Jan 2024, proving lasting relevance.

### Demand Signal for Two-Image Comparison

- On dive's original HN thread (2018), a user asked: "Is there a tool to compare 2 or more images to check which layers they have in common?"
- On dive's 2024 HN thread, users recommended container-diff and dredge as alternatives when comparison was needed.
- Google built container-diff in 2017, showing big-company validation of the use case. It was archived in 2024 only because Google couldn't maintain it, not because demand disappeared.
- The `docker scout compare` command (experimental) shows Docker Inc. recognizes the need, but their implementation is CVE-focused, not file-focused.

### Use Cases Driving Demand

1. **"Why did my image grow by 200MB?"** -- Most common: compare before/after a Dockerfile change
2. **"What changed between v1.2.3 and v1.2.4?"** -- Release validation, especially in regulated environments
3. **"Is this base image update safe?"** -- Comparing alpine:3.18 vs alpine:3.19
4. **CI/CD gating** -- Fail builds if image grows beyond threshold or new SUID binaries appear
5. **Supply chain security** -- Detect unexpected changes in third-party base images
6. **Incident response** -- "What changed in the image we just deployed?"

---

## 5. Distribution & Go-to-Market

### Channel Strategy (Ordered by Impact for Stars)

| Channel | Action | Expected Impact |
|---------|--------|-----------------|
| **Hacker News (Show HN)** | Launch post with GIF showing beautiful diff output | High. Dive got 684 pts on first Show HN. Container tools consistently do well. |
| **README with GIF/screenshot** | The most important asset. Make the diff output visually stunning. | Critical. This IS the marketing. |
| **Homebrew** | `brew install imgdiff` | Medium. Expected for Go CLI tools. |
| **GitHub Action** | `uses: yourname/imgdiff-action@v1` | High. Gives the tool a second life in CI. Every repo using it = a potential star. |
| **Docker plugin** | `docker imgdiff` | Medium-High. Puts it in Docker's native workflow. |
| **r/docker, r/devops, r/golang** | Cross-post launch | Medium. |
| **Dev.to / Hashnode blog posts** | "How I found a security issue in my base image using imgdiff" | Medium. Content-driven discovery. |
| **KubeCon / CloudNativeCon** | Lightning talk or project booth | Low-medium (but good for credibility). |
| **Twitter/X, BlueSky, Mastodon** | Share screenshots of interesting diffs | Ongoing. Container security finds are inherently shareable. |
| **Awesome lists** | awesome-docker, awesome-go, awesome-security | Medium. Evergreen discovery channel. |

### Name Considerations

- `imgdiff` -- short, memorable, available? Check npm/brew/GitHub.
- `cdiff` -- "container diff" but may conflict.
- `imdiff` -- image diff.
- `layerdiff` -- descriptive.
- `ocidiff` -- technical, correct.

**Recommendation:** Check GitHub/Homebrew availability. A short, memorable name matters more than a descriptive one.

### Launch Playbook

1. **Pre-launch (1 week before):** Seed 50-100 stars from personal network. A repo with 0 stars gets ignored.
2. **Launch day:** Show HN + Reddit (r/docker, r/devops, r/golang) + Twitter/X thread with GIF.
3. **Week 2:** Dev.to blog post: "I built git diff for container images."
4. **Week 3:** GitHub Action release. Blog post: "Add container image diffing to your CI pipeline in 5 minutes."
5. **Month 2:** Docker plugin. Interactive TUI mode. Second round of social posts.
6. **Ongoing:** Find interesting diffs in popular images and share them. "TIL alpine:3.20 added 47 new binaries vs 3.19" -- these posts go viral.

---

## 6. Founder-Idea Fit

| Dimension | Assessment | Score |
|-----------|-----------|-------|
| **Technical skill match** | Go experience (2 repos), deep Docker/K8s knowledge from running infra at fintech startup. This is exactly the right background. | 9/10 |
| **Domain knowledge** | Owns all infra at a fintech startup. Understands the pain of image bloat, security concerns, and CI/CD pipelines firsthand. | 9/10 |
| **Time availability** | Solo dev, side-project time only. MVP is 4-6 weeks of part-time work, which is achievable. Go tooling compiles fast and has excellent stdlib. | 7/10 |
| **Motivation alignment** | Optimizing for stars/reputation, not revenue. A CLI tool with beautiful output is ideal for this -- it's screenshottable, shareable, and compounds reputation. | 9/10 |
| **Ecosystem familiarity** | Uses Docker daily, understands OCI concepts, knows the tooling landscape. | 9/10 |

**Overall Founder-Idea Fit: 8.5/10** -- This is one of the best-fitted ideas for this founder profile. The only concern is time: maintaining an open-source tool requires ongoing effort (issues, PRs, releases), and dive's maintainer struggles even show that this can be a challenge.

---

## 7. Blind Spots & Contrarian Takes

### Blind Spot 1: "Most Developers Don't Compare Images"

The uncomfortable truth: most developers build an image, push it, and never look at what changed. The audience for this tool is primarily:
- Platform/DevOps engineers (not application developers)
- Security-conscious teams
- People debugging image size issues (episodic, not daily)

This means the tool might be used infrequently but valued highly when needed -- similar to dive's usage pattern ("saved my ass so many times" per HN comments). Infrequent use can still generate stars if the tool is impressive enough to bookmark.

### Blind Spot 2: "Docker Scout Is Coming for This"

Docker Inc. has `docker scout compare` as an experimental command. If Docker expands this to include file-level diffs (not just CVEs), it would significantly undercut a standalone tool. However:
- Docker Scout requires authentication and has rate limits on the free tier.
- Docker moves slowly on CLI features (the command has been "experimental" for 2+ years).
- A dedicated tool can iterate faster and go deeper.

### Blind Spot 3: "Dive Could Add This Feature Tomorrow"

Wagoodman has the libraries (stereoscope), the userbase (53K), and the domain expertise. If dive adds a `dive compare image1 image2` command, it would be hard to compete. However:
- Dive's maintainer is stretched thin (acknowledged in issue #568) and focused on stabilizing the existing tool.
- Dive's architecture is built around single-image TUI exploration, not two-image comparison.
- Anchore's priorities are security tools (syft/grype), not image diffing.

### Contrarian Take 1: "This Is a Feature, Not a Product"

Conventional wisdom says container image diff is "just a feature" that should live inside dive or Docker. The contrarian take: **the best developer tools do one thing exceptionally well.** `jq` is "just JSON filtering." `ripgrep` is "just grep." `bat` is "just cat." These all have 10K+ stars. A tool that does container image diff *perfectly* -- with beautiful output, security awareness, and CI integration -- can absolutely stand alone.

### Contrarian Take 2: "The Security Angle Is Underexplored"

Most container diff discussions focus on size optimization. But the *security* angle is where the real urgency lives:
- "Did someone add a new SUID binary to our production image?"
- "Did the base image update change file permissions?"
- "Are there new executable files that shouldn't be there?"

Framing this as a security tool (not just a diff tool) could unlock a different, more motivated audience.

### Contrarian Take 3: "AI Will Make This More Valuable, Not Less"

As AI generates more Dockerfiles and image configurations, the need to verify *what actually changed* increases. Developers will trust-but-verify AI-generated changes, and a diff tool becomes the verification layer.

---

## 8. Risk Matrix

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|------------|
| 1 | **dive adds comparison feature** | Low-Medium (20-30%) | High | Ship fast, build community, establish the tool as the "git diff for containers" before dive can react. Dive's slow development pace (one major release per year) gives you a window. |
| 2 | **Docker Scout expands to file-level diff** | Low (10-15%) | Very High | Docker moves slowly on CLI features. Position your tool as the open-source, no-auth-required alternative. |
| 3 | **Low adoption / "feature not product" perception** | Medium (30-40%) | Medium | Counter with exceptional UX, CI integration, and security framing. If the README GIF doesn't make people say "I need this," iterate on the output format. |
| 4 | **Maintenance burden exceeds side-project capacity** | Medium (30%) | Medium | Keep scope small. Accept community PRs. Establish co-maintainers early. Use GitHub Actions for release automation. |
| 5 | **Performance issues with large images** | Low-Medium (20%) | Medium | Implement manifest-only quick mode, streaming layer comparison, and shared-layer skipping from day one. |
| 6 | **Name collision / branding confusion** | Low (10%) | Low | Check GitHub, Homebrew, npm, and PyPI for name availability before committing. |
| 7 | **Security tooling companies build this into their platforms** | Medium (25%) | Low-Medium | Enterprise tools are paid and heavy. An open-source CLI tool coexists well -- see how trivy/grype coexist with commercial scanners. |
| 8 | **Saturated container tooling mindshare** | Medium (25%) | Medium | The container space has many tools, making discovery harder. Counter with exceptional README, active social presence, and consistent content. |

---

## 9. Recommendations & Next Steps

### Immediate Actions (Week 1)

1. **Validate the name.** Search GitHub, Homebrew, and Go packages for your preferred name. `imgdiff` and `ocicomp` are worth checking.
2. **Spike the core engine.** Use `go-containerregistry` to pull two images and iterate through their layer tars. Build the file tree data structure. This will validate feasibility and expose edge cases within 2-3 hours.
3. **Design the output format first.** Before writing diff logic, mock up the terminal output in a screenshot. This is your marketing material. Make it beautiful from day one. Study dive's TUI, `delta` (git diff beautifier), and `difftastic` for visual inspiration.

### MVP Priorities (Weeks 2-6)

4. **Core diff engine** with file-level comparison (added/removed/modified/permissions changed).
5. **Size impact summary** (total delta, largest additions, largest removals).
6. **Security highlights** (new SUID/SGID files, new executables, permission changes, new world-writable files).
7. **JSON output** for CI/CD integration.
8. **Color-coded terminal output** with lipgloss.
9. **Support local images** (docker daemon) and **remote images** (registries with auth).

### Post-MVP Priorities (Months 2-3)

10. **GitHub Action** -- this is the single highest-leverage distribution channel after the initial launch.
11. **Homebrew formula.**
12. **Interactive TUI mode** for browsing diffs (bubbletea).
13. **Package-level awareness** (detect apt/apk/pip package changes, not just file changes).
14. **`--quick` manifest-only mode** for instant registry comparisons.

### What NOT to Build

- Do not build a web UI or dashboard.
- Do not build vulnerability scanning (use syft/grype/trivy for that).
- Do not build image optimization/slimming (SlimToolkit does that).
- Do not build single-image exploration (dive does that).
- Stay focused on the *diff between two images*. That is the wedge.

### Star Target Estimates

| Milestone | Timeline | Confidence |
|-----------|----------|------------|
| 100 stars | Launch week | 85% (with proper Show HN + seeding) |
| 500 stars | Month 1 | 65% |
| 1,000 stars | Month 2-3 | 50% |
| 5,000 stars | Month 6-12 | 30% |
| 10,000+ stars | Year 1-2 | 15% |

The jump from 1K to 5K requires sustained content creation, CI integration adoption, and ideally a second viral moment (e.g., finding a security issue in a popular base image using your tool and writing about it).

### The One Thing That Matters Most

**The README GIF.** If someone lands on the repo and the GIF shows a beautiful, color-coded diff of two images with clear security callouts and size impact -- they will star it immediately. Invest disproportionate time in making the output format beautiful. Dive got 53K stars primarily because its TUI was immediately impressive in screenshots. Your tool needs that same "wow" moment, but for two-image comparison.

---

## Sources

### Competitor Repositories
- [wagoodman/dive](https://github.com/wagoodman/dive) -- 53.5K stars, Go, single-image layer explorer
- [GoogleContainerTools/container-diff](https://github.com/GoogleContainerTools/container-diff) -- 3.8K stars, archived March 2024
- [reproducible-containers/diffoci](https://github.com/reproducible-containers/diffoci) -- 561 stars, Go, semantic image diff
- [mthalman/dredge](https://github.com/mthalman/dredge) -- 44 stars, C#, registry client with image comparison
- [moul/docker-diff](https://github.com/moul/docker-diff) -- 157 stars, Shell, abandoned 2016
- [slimtoolkit/slim](https://github.com/slimtoolkit/slim) -- 23K stars, Go, image optimization (comparison is SaaS-only)
- [google/go-containerregistry](https://github.com/google/go-containerregistry) -- 3.8K stars, the standard Go library for container registries
- [anchore/stereoscope](https://github.com/anchore/stereoscope) -- Go library for layer file trees (powers dive and syft)

### Docker & Market Data
- [Docker Index: Continued Massive Developer Adoption](https://www.docker.com/blog/docker-index-shows-continued-massive-developer-adoption-and-activity-to-build-and-share-apps-with-docker/)
- [2025 Docker State of App Dev](https://www.docker.com/blog/2025-docker-state-of-app-dev/)
- [Docker Statistics and Facts 2025](https://expandedramblings.com/index.php/docker-statistics-facts/)
- [Container Security Market Size (MarketsandMarkets)](https://www.marketsandmarkets.com/PressReleases/container-security.asp)
- [Docker Scout Compare Docs](https://docs.docker.com/reference/cli/docker/scout/compare/)
- [Docker Container Diff Docs](https://docs.docker.com/reference/cli/docker/container/diff/)

### Hacker News Discussions
- [Dive HN Launch (Nov 2018, 684 points)](https://news.ycombinator.com/item?id=18528423)
- [Dive HN Discussion (Jan 2024, 464 points)](https://news.ycombinator.com/item?id=38913425)
- [Google Container-Diff HN (Nov 2017, 2 points)](https://news.ycombinator.com/item?id=15746505)

### Technical References
- [OCI Image Spec: Layer Filesystem Changeset](https://github.com/opencontainers/image-spec/blob/main/layer.md)
- [Interpreting Whiteout Files in Docker Image Layers](https://www.madebymikal.com/interpreting-whiteout-files-in-docker-image-layers/)
- [Docker Multi-Platform Builds](https://docs.docker.com/build/building/multi-platform/)
- [Tools for Analyzing Container Images (Jack Henschel)](https://blog.cubieserver.de/2024/tools-for-analyzing-and-working-with-container-images/)
- [Docker Image Analysis and Diffing (Augmented Mind)](https://www.augmentedmind.de/2023/02/05/docker-image-analysis-and-diffing/)
- [Dive Issue #568: Is This Project Still Alive?](https://github.com/wagoodman/dive/issues/568)

### Growth & Marketing
- [The Playbook for Getting More GitHub Stars](https://www.star-history.com/blog/playbook-for-more-github-stars)
- [How to Market on Hacker News (Tailscale Learnings)](https://www.markepear.dev/blog/developer-marketing-hacker-news)
- [Charmbracelet Bubbletea](https://github.com/charmbracelet/bubbletea) -- TUI framework for beautiful terminal output
- [Charmbracelet Lipgloss](https://github.com/charmbracelet/lipgloss) -- Terminal styling library
