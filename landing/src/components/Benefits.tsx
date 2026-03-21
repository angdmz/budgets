export function Benefits() {
  const benefits = [
    {
      title: "Save More Money",
      description: "Users save an average of 20% more when actively tracking their budgets.",
      stat: "20%",
    },
    {
      title: "Reduce Stress",
      description: "Know exactly where your money is going and eliminate financial anxiety.",
      stat: "85%",
    },
    {
      title: "Achieve Goals",
      description: "Set and reach your financial goals faster with clear visibility.",
      stat: "3x",
    },
  ];

  return (
    <section className="py-20 px-4 sm:px-6 lg:px-8 bg-gray-50">
      <div className="max-w-7xl mx-auto">
        <div className="text-center mb-16">
          <h2 className="text-3xl sm:text-4xl font-bold text-gray-900 mb-4">
            Why Choose Budget Manager?
          </h2>
          <p className="text-xl text-gray-600 max-w-2xl mx-auto">
            Join thousands of users who have transformed their financial lives.
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
