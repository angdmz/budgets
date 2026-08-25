import { getDictionary } from '@/i18n/getDictionary';
import { locales, defaultLocale, type Locale } from '@/i18n/config';
import type { Metadata } from 'next';
import { Inter } from 'next/font/google';

const inter = Inter({ subsets: ['latin'] });

const BASE_URL = process.env.NEXT_PUBLIC_BASE_URL || 'https://yourdomain.com';

export async function generateMetadata({ params }: { params: { lang: Locale } }): Promise<Metadata> {
  const dict = await getDictionary(params.lang);

  return {
    title: dict.metadata.title,
    description: dict.metadata.description,
    openGraph: {
      title: dict.metadata.ogTitle,
      description: dict.metadata.ogDescription,
      type: 'website',
      locale: params.lang,
    },
    twitter: {
      card: 'summary_large_image',
      title: dict.metadata.twitterTitle,
      description: dict.metadata.twitterDescription,
    },
    alternates: {
      canonical: `${BASE_URL}/${params.lang}`,
      languages: {
        ...Object.fromEntries(locales.map(l => [l, `${BASE_URL}/${l}`])),
        'x-default': `${BASE_URL}/${defaultLocale}`,
      },
    },
    robots: { index: true, follow: true },
  };
}

export function generateStaticParams() {
  return locales.map(lang => ({ lang }));
}

export default function LocaleLayout({
  children,
  params,
}: {
  children: React.ReactNode;
  params: { lang: Locale };
}) {
  return (
    <html lang={params.lang}>
      <body className={inter.className}>{children}</body>
    </html>
  );
}
