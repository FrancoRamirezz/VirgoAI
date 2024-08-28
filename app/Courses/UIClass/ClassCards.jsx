import React from "react"

import React from 'react';

const CourseCard = ({ title, description, icon }) => (
  <div className="bg-white bg-opacity-90 rounded-lg shadow-md mb-6 hover:shadow-xl transition-all duration-300 transform hover:-translate-y-1 overflow-hidden w-full max-w-md">
    <div className="p-6">
      <div className="flex items-center mb-4">
        <span className="text-3xl mr-4">{icon}</span>
        <h2 className="text-xl font-semibold text-purple-800">{title}</h2>
      </div>
      <p className="text-gray-700">{description}</p>
    </div>
  </div>
);

const CoursePortal = () => {
  const courses = [
    { 
      title: "English 101", 
      description: "Introduction to English composition and literature. Develop your writing skills and explore classic works.",
      icon: "📚"
    },
    { 
      title: "Spanish to English", 
      description: "Learn to translate from Spanish to English. Enhance your bilingual skills and cultural understanding.",
      icon: "🇪🇸🔄🇬🇧"
    },
    { 
      title: "History of English", 
      description: "Explore the origins and evolution of the English language from Old English to modern times.",
      icon: "🏛️"
    }
  ];

  return (
    <div className="min-h-screen bg-gradient-to-b from-purple-100 to-blue-100 flex flex-col justify-center items-center p-8 text-center">
      <h1 className="text-4xl font-bold mb-8 text-purple-900">Language Courses</h1>
      <div className="w-full max-w-2xl">
        {courses.map((course, index) => (
          <CourseCard key={index} title={course.title} description={course.description} icon={course.icon} />
        ))}
      </div>
    </div>
  );
};

export default CourseCard;