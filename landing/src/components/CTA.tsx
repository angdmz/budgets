import Link from 'next/link';

export function CTA() {
  return (
    <section className="py-20 px-4 sm:px-6 lg:px-8 bg-primary-600">
      <div className="max-w-4xl mx-auto text-center">
        <h2 className="text-3xl sm:text-4xl font-bold text-white mb-6">
          Ready to Take Control of Your Finances?
        </h2>
        <p className="text-xl text-primary-100 mb-8">
          Start managing your budget today. No credit card required.
        </p>
        <Link
          href="/app"
          className="inline-block bg-white text-primary-600 hover:bg-gray-100 px-8 py-4 rounded-lg text-lg font-semibold shadow-lg hover:shadow-xl transition-all"
        >
          Get Started Free
        </Link>
      </div>
    </section>
  );
}
