import Link from 'next/link';

export function Hero() {
  return (
    <section className="pt-24 pb-16 px-4 sm:px-6 lg:px-8 bg-gradient-to-b from-primary-50 to-white">
      <div className="max-w-7xl mx-auto">
        <div className="text-center">
          <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold text-gray-900 mb-6">
            Take Control of Your
            <span className="text-primary-600"> Finances</span>
          </h1>
          <p className="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">
            A powerful budget management system with multi-user support, group-based budgeting, 
            and real-time expense tracking. Start managing your finances better today.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link
              href="/app"
              className="bg-primary-600 text-white hover:bg-primary-700 px-8 py-3 rounded-lg text-lg font-semibold shadow-lg hover:shadow-xl transition-all"
            >
              Get Started Free
            </Link>
            <Link
              href="#features"
              className="bg-white text-primary-600 hover:bg-gray-50 px-8 py-3 rounded-lg text-lg font-semibold border-2 border-primary-600 transition-all"
            >
              Learn More
            </Link>
          </div>
        </div>
        <div className="mt-16 relative">
          <div className="bg-white rounded-lg shadow-2xl p-8 max-w-4xl mx-auto">
            <div className="aspect-video bg-gradient-to-br from-primary-100 to-primary-200 rounded-lg flex items-center justify-center">
              <p className="text-primary-700 text-2xl font-semibold">Dashboard Preview</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
