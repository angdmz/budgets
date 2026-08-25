import { getDictionary } from '@/i18n/getDictionary';
import type { Locale } from '@/i18n/config';
import { Hero } from '@/components/Hero';
import { Features } from '@/components/Features';
import { Benefits } from '@/components/Benefits';
import { CTA } from '@/components/CTA';
import { Footer } from '@/components/Footer';
import { Navigation } from '@/components/Navigation';

export default async function Home({ params }: { params: { lang: Locale } }) {
  const dict = await getDictionary(params.lang);
  
  return (
    <main className="min-h-screen">
      <Navigation dict={dict.nav} lang={params.lang} />
      <Hero dict={dict.hero} />
      <Features dict={dict.features} />
      <Benefits dict={dict.benefits} />
      <CTA dict={dict.cta} />
      <Footer dict={dict.footer} />
    </main>
  );
}
