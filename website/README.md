# Nestor Website

This directory contains the static website for the Nestor NES emulator, built with Hugo.

## Development

### Local Development

```bash
# Install Hugo (if not already installed)
# macOS: brew install hugo
# Linux: apt install hugo / snap install hugo
# Windows: choco install hugo

# Run development server
hugo server
```

The site will be available at http://localhost:1313

### Building

```bash
# Build static files
hugo --minify --cleanDestinationDir
```

The built site will be in the `public/` directory.

## GitHub Pages Deployment

This site is designed to be deployed on GitHub Pages. The workflow is typically:

1. Make changes to content or theme
2. Test locally with `hugo server`
3. Commit changes to the repository
4. GitHub Pages will automatically build and deploy using Hugo

## Structure

```
website/
├── config.yaml              # Hugo configuration
├── content/                  # Markdown content files
│   ├── _index.md            # Home page
│   ├── features.md          # Features page
│   ├── mappers.md           # Mappers page
│   ├── screenshots.md       # Screenshots page
│   ├── docs.md              # Documentation page
│   └── about.md             # About page
├── static/                   # Static assets
│   └── images/
│       └── logo.png         # Nestor logo
└── themes/nestor-theme/     # Custom Hugo theme
    ├── layouts/
    │   ├── _default/
    │   │   ├── baseof.html  # Base template
    │   │   └── single.html  # Single page template
    │   └── index.html       # Home page template
    └── static/css/
        └── style.css        # Theme styles
```

## Editing Content

Content is written in Markdown and stored in the `content/` directory. Each page has front matter with title and description for SEO.

## Theme

The site uses a custom Hugo theme called "nestor-theme" with simple, clean styling appropriate for a hobby project.
│   └── static/css/
│       └── style.css        # Main stylesheet
├── Dockerfile               # Multi-stage Docker build
├── docker-compose.yml       # Development and production setup
├── nginx.conf               # Production nginx configuration
└── README.md               # This file
```

## Features

- **Responsive Design**: Works on desktop and mobile
- **Fast Loading**: Optimized assets and minimal dependencies
- **SEO Friendly**: Meta tags and structured data
- **Accessible**: WCAG compliant design
- **Modern CSS**: CSS Grid and Flexbox layouts
- **Docker Ready**: Containerized for easy deployment

## Theme Customization

The custom `nestor-theme` includes:

- Clean, modern design inspired by retro gaming
- NES-inspired color palette
- Mobile-first responsive layout
- Semantic HTML structure
- Optimized CSS with CSS custom properties
- Interactive navigation with mobile menu

## Content Management

All content is written in Markdown with YAML frontmatter:

```yaml
---
title: "Page Title"
description: "Page description for SEO"
---

# Page Content

Your markdown content here...
```

## Deployment

### GitHub Pages

1. Build the site: `hugo --minify --cleanDestinationDir`
2. Deploy the `public/` directory to GitHub Pages
3. Configure custom domain if needed

### Docker Deployment

```bash
# Build and run
docker build -t nestor-website .
docker run -p 3000:80 nestor-website
```

### Traditional Web Server

1. Build the site: `hugo --minify --cleanDestinationDir`
2. Copy `public/` directory contents to web root
3. Configure web server for clean URLs

## Development

### Adding New Pages

1. Create a new `.md` file in `content/`
2. Add frontmatter with title and description
3. Write content in Markdown
4. Add to navigation menu in `config.yaml` if needed

### Updating Styles

Edit `themes/nestor-theme/static/css/style.css` for theme changes.

### Adding Images

Place images in `static/images/` and reference with:
```markdown
![Alt text](images/filename.png)
```

## Performance

The website is optimized for performance:

- **Hugo**: Fast static site generation
- **Nginx**: Efficient web server with compression
- **Minimal CSS**: Single stylesheet, no frameworks
- **Optimized Images**: Compressed assets
- **CDN Ready**: Static files suitable for CDN deployment

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Mobile browsers (iOS Safari, Chrome Mobile)
- Graceful degradation for older browsers

## License

The website content and theme are part of the Nestor project and follow the same GPL v3 license.