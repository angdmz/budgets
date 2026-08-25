import Link from 'next/link';

interface NavigationProps {
  dict: {
    brand: string;
    signIn: string;
    getStarted: string;
  };
  lang: string;
}

export function Navigation({ dict, lang }: NavigationProps) {
  return (
    <nav className="fixed w-full bg-white/90 backdrop-blur-sm shadow-sm z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          <div className="flex items-center">
            <Link href={`/${lang}`} className="text-2xl font-bold text-primary-600">
              {dict.brand}
            </Link>
          </div>
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2 mr-2">
              <Link 
                href="/en" 
                className={`px-2 py-1 rounded text-sm font-medium transition-colors ${
                  lang === 'en' 
                    ? 'text-primary-600 font-bold bg-primary-50' 
                    : 'text-gray-500 hover:text-primary-600'
                }`}
              >
                EN
              </Link>
              <span className="text-gray-300">|</span>
              <Link 
                href="/es" 
                className={`px-2 py-1 rounded text-sm font-medium transition-colors ${
                  lang === 'es' 
                    ? 'text-primary-600 font-bold bg-primary-50' 
                    : 'text-gray-500 hover:text-primary-600'
                }`}
              >
                ES
              </Link>
            </div>
            <Link
              href="/app"
              className="text-gray-700 hover:text-primary-600 px-3 py-2 rounded-md text-sm font-medium"
            >
              {dict.signIn}
            </Link>
            <Link
              href="/app"
              className="bg-primary-600 text-white hover:bg-primary-700 px-4 py-2 rounded-md text-sm font-medium"
            >
              {dict.getStarted}
            </Link>
          </div>
        </div>
      </div>
    </nav>
  );
}
