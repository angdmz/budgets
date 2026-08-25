interface BenefitsProps {
  dict: {
    sectionTitle: string;
    sectionSubtitle: string;
    saveMoney: { stat: string; title: string; description: string };
    reduceStress: { stat: string; title: string; description: string };
    achieveGoals: { stat: string; title: string; description: string };
  };
}

export function Benefits({ dict }: BenefitsProps) {
  const benefits = [
    {
      title: dict.saveMoney.title,
      description: dict.saveMoney.description,
      stat: dict.saveMoney.stat,
    },
    {
      title: dict.reduceStress.title,
      description: dict.reduceStress.description,
      stat: dict.reduceStress.stat,
    },
    {
      title: dict.achieveGoals.title,
      description: dict.achieveGoals.description,
      stat: dict.achieveGoals.stat,
    },
  ];

  return (
    <section className="py-20 px-4 sm:px-6 lg:px-8 bg-gray-50">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
            {dict.sectionTitle}
          </h2>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            {dict.sectionSubtitle}
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {benefits.map((benefit, index) => (
            <div key={index} className="bg-white p-8 rounded-lg shadow-md text-center">
              <div className="text-5xl font-bold text-primary-600 mb-4">
                {benefit.stat}
              </div>
              <h3 className="text-xl font-semibold text-gray-900 mb-2">
                {benefit.title}
              </h3>
              <p className="text-gray-600">{benefit.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
