import { MetadataRoute } from 'next';
import { locales } from '@/i18n/config';

const BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'https://yourdomain.com';

export default function sitemap(): MetadataRoute.Sitemap {
  return locales.map(lang => ({
    url: `${BASE_URL}/${lang}`,
    lastModified: new Date(),
    changeFrequency: 'monthly' as const,
    priority: 1,
    alternates: {
      languages: Object.fromEntries(locales.map(l => [l, `${BASE_URL}/${l}`])),
    },
  }));
}
