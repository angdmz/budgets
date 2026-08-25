import Link from 'next/link';

interface FooterProps {
  dict: {
    brand: string;
    tagline: string;
    product: string;
    features: string;
    pricing: string;
    demo: string;
    company: string;
    about: string;
    contact: string;
    privacy: string;
    support: string;
    helpCenter: string;
    documentation: string;
    api: string;
    copyright: string;
  };
}

export function Footer({ dict }: FooterProps) {
  return (
    <footer className="bg-gray-900 text-gray-300 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-8">
          <div>
            <h3 className="text-white text-lg font-semibold mb-4">{dict.brand}</h3>
            <p className="text-sm">
              {dict.tagline}
            </p>
          </div>
          <div>
            <h4 className="text-white font-semibold mb-4">{dict.product}</h4>
            <ul className="space-y-2 text-sm">
              <li><Link href="#features" className="hover:text-white">{dict.features}</Link></li>
              <li><Link href="/app" className="hover:text-white">{dict.pricing}</Link></li>
              <li><Link href="/app" className="hover:text-white">{dict.demo}</Link></li>
            </ul>
          </div>
          <div>
            <h4 className="text-white font-semibold mb-4">{dict.company}</h4>
            <ul className="space-y-2 text-sm">
              <li><Link href="/about" className="hover:text-white">{dict.about}</Link></li>
              <li><Link href="/contact" className="hover:text-white">{dict.contact}</Link></li>
              <li><Link href="/privacy" className="hover:text-white">{dict.privacy}</Link></li>
            </ul>
          </div>
          <div>
            <h4 className="text-white font-semibold mb-4">{dict.support}</h4>
            <ul className="space-y-2 text-sm">
              <li><Link href="/help" className="hover:text-white">{dict.helpCenter}</Link></li>
              <li><Link href="/docs" className="hover:text-white">{dict.documentation}</Link></li>
              <li><Link href="/api" className="hover:text-white">{dict.api}</Link></li>
            </ul>
          </div>
        </div>
        <div className="border-t border-gray-800 pt-8 text-sm text-center">
          <p>{dict.copyright.replace('{{year}}', new Date().getFullYear().toString())}</p>
        </div>
      </div>
    </footer>
  );
}
