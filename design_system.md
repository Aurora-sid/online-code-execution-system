## Design System: Aurora Code

### Pattern
- **Name:** App Store Style Landing
- **Conversion Focus:** Show real screenshots. Include ratings (4.5+ stars). QR code for mobile. Platform-specific CTAs.
- **CTA Placement:** Download buttons prominent (App Store + Play Store) throughout
- **Color Strategy:** Dark/light matching app store feel. Star ratings in gold. Screenshots with device frames.
- **Sections:** 1. Hero with device mockup, 2. Screenshots carousel, 3. Features with icons, 4. Reviews/ratings, 5. Download CTAs

### Style
- **Name:** Vibrant & Block-based
- **Keywords:** Bold, energetic, playful, block layout, geometric shapes, high color contrast, duotone, modern, energetic
- **Best For:** Startups, creative agencies, gaming, social media, youth-focused, entertainment, consumer
- **Performance:** 鈿?Good | **Accessibility:** 鈼?Ensure WCAG

### Colors
| Role | Hex |
|------|-----|
| Primary | #3B82F6 |
| Secondary | #1E293B |
| CTA | #2563EB |
| Background | #0F172A |
| Text | #F1F5F9 |

*Notes: Dark syntax theme colors + Blue focus*

### Typography
- **Heading:** JetBrains Mono
- **Body:** IBM Plex Sans
- **Mood:** code, developer, technical, precise, functional, hacker
- **Best For:** Developer tools, documentation, code editors, tech blogs, CLI apps
- **Google Fonts:** https://fonts.google.com/share?selection.family=IBM+Plex+Sans:wght@300;400;500;600;700|JetBrains+Mono:wght@400;500;600;700
- **CSS Import:**
```css
@import url('https://fonts.googleapis.com/css2?family=IBM+Plex+Sans:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500;600;700&display=swap');
```

### Key Effects
Large sections (48px+ gaps), animated patterns, bold hover (color shift), scroll-snap, large type (32px+), 200-300ms

### Avoid (Anti-patterns)
- Flat design without depth
- Text-heavy pages

### Pre-Delivery Checklist
- [ ] No emojis as icons (use SVG: Heroicons/Lucide)
- [ ] cursor-pointer on all clickable elements
- [ ] Hover states with smooth transitions (150-300ms)
- [ ] Light mode: text contrast 4.5:1 minimum
- [ ] Focus states visible for keyboard nav
- [ ] prefers-reduced-motion respected
- [ ] Responsive: 375px, 768px, 1024px, 1440px

