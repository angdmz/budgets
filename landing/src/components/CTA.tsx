import Link from 'next/link';

interface CTAProps {
  dict: {
    title: string;
    subtitle: string;
    button: string;
  };
}

export function CTA({ dict }: CTAProps) {
  return (
    <section className="py-20 px-4 sm:px-6 lg:px-8 bg-primary-600">
      <div className="max-w-4xl mx-auto text-center">
        <h2 className="text-3xl sm:text-4xl font-bold text-white mb-6">
          {dict.title}
        </h2>
        <p className="text-xl text-primary-100 mb-8">
          {dict.subtitle}
        </p>
        <Link
          href="/app"
          className="inline-block bg-white text-primary-600 hover:bg-gray-100 px-8 py-4 rounded-lg text-lg font-semibold shadow-lg hover:shadow-xl transition-all"
        >
          {dict.button}
        </Link>
      </div>
    </section>
  );
}
