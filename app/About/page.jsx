import React from 'react';
import { ArrowRight, BookOpen, Globe, Users } from 'lucide-react';

const AboutPage = () => {
  return (
    <div className="min-h-screen bg-gradient-to-b from-purple-100 to-blue-100 p-8">
      <header className="mb-12 text-center">
        <h1 className="text-4xl font-bold text-purple-600 mb-4">About Us</h1>
        <p className="text-xl text-gray-700">Empowering Your Journey to Citizenship</p>
      </header>

      <main className="max-w-4xl mx-auto">
        <section className="bg-white rounded-lg shadow-md p-8 mb-12">
          <h2 className="text-2xl font-semibold text-purple-800 mb-4">Our Mission</h2>
          <p className="text-gray-600 mb-4">
            At [Your LMS Name], our goal is to help people learn English effectively, 
            focusing on the skills needed to pass the citizenship test. We specialize in 
            assisting Spanish and Chinese speakers in their journey to becoming 
            confident English communicators and successful U.S. citizens.
          </p>
          <p className="text-gray-600">
            We believe that language should never be a barrier to achieving your dreams. 
            Our dedicated team works tirelessly to create engaging, accessible, and 
            effective learning materials tailored to your needs.
          </p>
        </section>

        <section className="mb-12">
          <h2 className="text-2xl font-semibold text-purple-800 mb-6">What We Offer</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {[
              { 
                title: 'Tailored Curriculum', 
                description: 'Courses designed specifically for the citizenship test',
                icon: <BookOpen className="w-6 h-6 text-purple-500" />
              },
              { 
                title: 'Language Support', 
                description: 'Specialized help for Spanish and Chinese speakers',
                icon: <Globe className="w-6 h-6 text-blue-500" />
              },
              { 
                title: 'Interactive Learning', 
                description: 'Engaging exercises to build your confidence',
                icon: <Users className="w-6 h-6 text-purple-500" />
              },
              { 
                title: 'Progress Tracking', 
                description: 'Monitor your improvement as you learn',
                icon: <ArrowRight className="w-6 h-6 text-blue-500" />
              }
            ].map((feature, index) => (
              <div key={index} className="bg-white rounded-lg shadow-md p-6 flex items-start">
                <div className="mr-4">
                  {feature.icon}
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-purple-700 mb-2">{feature.title}</h3>
                  <p className="text-gray-600">{feature.description}</p>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className="bg-white rounded-lg shadow-md p-8 text-center">
          <h2 className="text-2xl font-semibold text-purple-800 mb-4">Start Your Journey Today</h2>
          <p className="text-gray-600 mb-6">
            Join thousands of successful learners who have achieved their dreams of U.S. citizenship.
          </p>
          <button className="bg-purple-600 hover:bg-purple-700 text-white font-bold py-2 px-6 rounded-full transition duration-300 ease-in-out transform hover:scale-105">
            Begin Learning
          </button>
        </section>
      </main>
    </div>
  );
};

export default AboutPage;
