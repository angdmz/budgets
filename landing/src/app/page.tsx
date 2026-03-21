import Link from 'next/link';
import { Hero } from '@/components/Hero';
import { Features } from '@/components/Features';
import { Benefits } from '@/components/Benefits';
import { CTA } from '@/components/CTA';
import { Footer } from '@/components/Footer';
import { Navigation } from '@/components/Navigation';

export default function Home() {
  return (
    <main className="min-h-screen">
      <Navigation />
      <Hero />
      <Features />
      <Benefits />
      <CTA />
      <Footer />
    </main>
  );
}
