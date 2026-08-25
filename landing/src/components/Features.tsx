interface FeaturesProps {
  dict: {
    sectionTitle: string;
    sectionSubtitle: string;
    multiUser: { title: string; description: string };
    groupBudgeting: { title: string; description: string };
    realTime: { title: string; description: string };
    secure: { title: string; description: string };
    categories: { title: string; description: string };
    reports: { title: string; description: string };
  };
}

export function Features({ dict }: FeaturesProps) {
  const features = [
    {
      title: dict.multiUser.title,
      description: dict.multiUser.description,
      icon: "👥",
    },
    {
      title: dict.groupBudgeting.title,
      description: dict.groupBudgeting.description,
      icon: "📊",
    },
    {
      title: dict.realTime.title,
      description: dict.realTime.description,
      icon: "⚡",
    },
    {
      title: dict.secure.title,
      description: dict.secure.description,
      icon: "🔒",
    },
    {
      title: dict.categories.title,
      description: dict.categories.description,
      icon: "🏷️",
    },
    {
      title: dict.reports.title,
      description: dict.reports.description,
      icon: "📈",
    },
  ];

  return (
    <section id="features" className="py-20 px-4 sm:px-6 lg:px-8 bg-white">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
            {dict.sectionTitle}
          </h2>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            {dict.sectionSubtitle}
          </p>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {features.map((feature, index) => (
            <div
              key={index}
              className="bg-white p-6 rounded-lg border-2 border-gray-100 hover:border-primary-300 hover:shadow-lg transition-all"
            >
              <div className="text-4xl mb-4">{feature.icon}</div>
              <h3 className="text-xl font-semibold text-gray-900 mb-2">
                {feature.title}
              </h3>
              <p className="text-gray-600">{feature.description}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
