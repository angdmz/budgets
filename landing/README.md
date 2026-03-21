# Budget Management System - Landing Page

## Overview

SEO-optimized marketing landing page built with Next.js 14 using the App Router. Features server-side rendering for optimal performance and search engine visibility.

## Technology Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Rendering**: Server-Side Rendering (SSR) + Static Site Generation (SSG)
- **Deployment**: Docker (standalone output)

## Features

### SEO Optimization
- ✅ Meta tags (title, description, keywords)
- ✅ Open Graph tags for social sharing
- ✅ Twitter Card tags
- ✅ Sitemap generation (`/sitemap.xml`)
- ✅ Robots.txt configuration
- ✅ Semantic HTML structure
- ✅ Optimized images and assets

### Page Sections
- **Navigation** - Sticky header with CTA buttons
- **Hero** - Main value proposition with call-to-action
- **Features** - 6 key features with icons
- **Benefits** - Statistics and user benefits
- **CTA** - Final call-to-action section
- **Footer** - Links and company information

### Performance
- Server-side rendering for fast initial load
- Optimized images with Next.js Image component
- Minimal JavaScript bundle
- Lighthouse score target: 90+

## Setup & Development

### Prerequisites
- Node.js 20 or higher
- npm or yarn

### Installation

```bash
# Install dependencies
npm install
```

### Environment Variables

No environment variables required for the landing page. All configuration is static.

### Development Mode

```bash
# Start development server
npm run dev
```

Access at: `http://localhost:3000`

The development server includes:
- Hot module replacement
- Fast refresh
- Error overlay
- TypeScript checking

### Build for Production

```bash
# Create production build
npm run build

# Start production server
npm start
```

### Docker Build

```bash
# Build Docker image
docker build -t budget-landing .

# Run container
docker run -p 3000:3000 budget-landing
```

## Project Structure

```
landing/
├── src/
│   ├── app/
│   │   ├── layout.tsx        # Root layout with SEO metadata
│   │   ├── page.tsx          # Home page
│   │   ├── globals.css       # Global styles
│   │   └── sitemap.ts        # Sitemap generation
│   └── components/
│       ├── Navigation.tsx    # Header navigation
│       ├── Hero.tsx          # Hero section
│       ├── Features.tsx      # Features grid
│       ├── Benefits.tsx      # Benefits section
│       ├── CTA.tsx           # Call-to-action
│       └── Footer.tsx        # Footer links
├── public/
│   └── robots.txt            # Search engine directives
├── tailwind.config.ts        # Tailwind configuration
├── next.config.mjs           # Next.js configuration
├── tsconfig.json             # TypeScript configuration
├── package.json              # Dependencies
└── Dockerfile                # Production build
```

## Customization

### Updating Content

**Hero Section** (`src/components/Hero.tsx`):
```tsx
<h1>Your Custom Headline</h1>
<p>Your custom description</p>
```

**Features** (`src/components/Features.tsx`):
```tsx
const features = [
  {
    title: "Your Feature",
    description: "Feature description",
    icon: "🎯",
  },
  // Add more features
];
```

**SEO Metadata** (`src/app/layout.tsx`):
```tsx
export const metadata: Metadata = {
  title: "Your Title",
  description: "Your description",
  keywords: ["your", "keywords"],
};
```

### Styling

The landing page uses Tailwind CSS. Customize the theme in `tailwind.config.ts`:

```typescript
theme: {
  extend: {
    colors: {
      primary: {
        // Your color palette
      },
    },
  },
}
```

### Adding New Sections

1. Create component in `src/components/`
2. Import in `src/app/page.tsx`
3. Add to page layout

Example:
```tsx
// src/components/Testimonials.tsx
export function Testimonials() {
  return (
    <section className="py-20">
      {/* Your content */}
    </section>
  );
}

// src/app/page.tsx
import { Testimonials } from '@/components/Testimonials';

export default function Home() {
  return (
    <main>
      <Hero />
      <Features />
      <Testimonials /> {/* New section */}
      <CTA />
    </main>
  );
}
```

## SEO Best Practices

### Implemented
- ✅ Semantic HTML5 elements
- ✅ Proper heading hierarchy (h1 → h6)
- ✅ Alt text for images
- ✅ Meta descriptions under 160 characters
- ✅ Mobile-responsive design
- ✅ Fast page load times
- ✅ HTTPS ready (via Nginx)

### Recommendations
- Add structured data (JSON-LD)
- Implement analytics (Google Analytics, Plausible)
- Add social media integration
- Create blog section for content marketing
- Implement A/B testing

## Performance Optimization

### Current Optimizations
- Next.js automatic code splitting
- Image optimization with next/image
- CSS purging with Tailwind
- Minimal JavaScript bundle
- Server-side rendering

### Monitoring

Use Lighthouse to check performance:
```bash
npm run build
npm start
# Then run Lighthouse in Chrome DevTools
```

Target scores:
- Performance: 90+
- Accessibility: 95+
- Best Practices: 95+
- SEO: 100

## Deployment

### Docker Deployment

The landing page uses Next.js standalone output for minimal Docker image size.

**Build:**
```bash
docker build -t budget-landing:latest .
```

**Run:**
```bash
docker run -p 3000:3000 budget-landing:latest
```

### Production Checklist
- [ ] Update meta tags with production URLs
- [ ] Update sitemap with production domain
- [ ] Configure CDN for static assets
- [ ] Set up analytics
- [ ] Test on multiple devices/browsers
- [ ] Verify all links work
- [ ] Check mobile responsiveness
- [ ] Run Lighthouse audit
- [ ] Test page load speed
- [ ] Verify SEO metadata

## Integration with Main App

The landing page links to the main application:

```tsx
<Link href="/app">Get Started</Link>
```

When deployed with Nginx, this routes to the React application at `/app`.

## Troubleshooting

### Build Errors

**"Module not found"**
```bash
rm -rf node_modules .next
npm install
npm run build
```

**TypeScript errors**
```bash
npm run lint
# Fix reported issues
```

### Runtime Issues

**Styles not loading**
- Check Tailwind configuration
- Verify `globals.css` is imported in `layout.tsx`
- Clear `.next` cache

**Images not displaying**
- Ensure images are in `public/` directory
- Use correct paths (relative to public)
- Check Next.js Image component configuration

## Contributing

### Code Style
- Use TypeScript for type safety
- Follow Next.js conventions
- Use Tailwind for styling (no custom CSS)
- Keep components small and focused
- Write semantic HTML

### Testing
```bash
# Type checking
npm run build

# Linting
npm run lint
```

## License

See LICENSE file in repository root.
