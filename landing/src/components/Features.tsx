export function Features() {
  const features = [
    {
      title: "Multi-User Support",
      description: "Collaborate with family members or team members. Share budgets and track expenses together.",
      icon: "👥",
    },
    {
      title: "Group-Based Budgeting",
      description: "Create separate budgets for different groups - family, business, or personal projects.",
      icon: "📊",
    },
    {
      title: "Real-Time Tracking",
      description: "Track expected vs actual expenses in real-time. See where your money is going instantly.",
      icon: "⚡",
    },
    {
      title: "Secure & Encrypted",
      description: "Your financial data is encrypted at rest. Bank-level security for your peace of mind.",
      icon: "🔒",
    },
    {
      title: "Custom Categories",
      description: "Create custom expense categories that match your lifestyle and spending patterns.",
      icon: "🏷️",
    },
    {
      title: "Detailed Reports",
      description: "Generate comprehensive reports to understand your spending habits and make better decisions.",
      icon: "📈",
    },
  ];

  return (
    <section id="features" className="py-20 px-4 sm:px-6 lg:px-8 bg-white">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
            Everything You Need to Manage Your Budget
          </h2>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            Powerful features designed to make budget management simple and effective.
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
