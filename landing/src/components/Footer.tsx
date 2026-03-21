import Link from 'next/link';

export function Footer() {
  return (
    <footer className="bg-gray-900 text-gray-300 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-7xl mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-8">
          <div>
            <h3 className="text-white text-lg font-semibold mb-4">Budget Manager</h3>
            <p className="text-sm">
              Take control of your finances with our powerful budget management system.
            </p>
          </div>
          <div>
            <h4 className="text-white font-semibold mb-4">Product</h4>
            <ul className="space-y-2 text-sm">
              <li><Link href="#features" className="hover:text-white">Features</Link></li>
              <li><Link href="/app" className="hover:text-white">Pricing</Link></li>
              <li><Link href="/app" className="hover:text-white">Demo</Link></li>
            </ul>
          </div>
          <div>
            <h4 className="text-white font-semibold mb-4">Company</h4>
            <ul className="space-y-2 text-sm">
              <li><Link href="/about" className="hover:text-white">About</Link></li>
              <li><Link href="/contact" className="hover:text-white">Contact</Link></li>
              <li><Link href="/privacy" className="hover:text-white">Privacy</Link></li>
            </ul>
          </div>
          <div>
            <h4 className="text-white font-semibold mb-4">Support</h4>
            <ul className="space-y-2 text-sm">
              <li><Link href="/help" className="hover:text-white">Help Center</Link></li>
              <li><Link href="/docs" className="hover:text-white">Documentation</Link></li>
              <li><Link href="/api" className="hover:text-white">API</Link></li>
            </ul>
          </div>
        </div>
        <div className="border-t border-gray-800 pt-8 text-sm text-center">
          <p>&copy; {new Date().getFullYear()} Budget Manager. All rights reserved.</p>
        </div>
      </div>
    </footer>
  );
}
