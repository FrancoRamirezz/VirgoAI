"use client";

import React from "react";
import Image from "next/image";
import { TypeAnimation } from "react-type-animation";
import { motion } from "framer-motion";
import Link from "next/link";
import AiLogo from "/Users/franciscoramirez/Desktop/MLH/ml-hack/public/Images/AiLogo.jpg";

const HeroSection = () => {
  return (
    <div className="min-h-screen bg-gradient-to-b from-purple-100 to-blue-100 flex flex-col justify-center items-center p-8 text-center">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        className="mb-8"
      >
        <Image
          src={AiLogo}
          alt="AI Logo"
          width={150}
          height={150}
          className="rounded-full shadow-lg"
        />
      </motion.div>

      <motion.h1
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5, delay: 0.2 }}
        className="text-4xl font-bold text-purple-600 mb-4"
      >
        Hello, I'm{" "}
        <TypeAnimation
          sequence={[
            "AI Citizenship Test Helper",
            1000,
            "Your Study Buddy",
            1000,
            "Your Path to Success",
            1000,
          ]}
          wrapper="span"
          speed={50}
          repeat={Infinity}
          className="text-blue-600"
        />
      </motion.h1>

      <motion.p
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5, delay: 0.4 }}
        className="text-xl text-gray-700 mb-8"
      >
        Welcome to AI Citizenship Test Helper. 
        We're here to assist Spanish and Chinese speakers in learning English for the citizenship test.
      </motion.p>

      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5, delay: 0.6 }}
      >
        <Link href="/get-started" passHref>
          <button className="bg-purple-600 hover:bg-purple-700 text-white font-bold py-3 px-6 rounded-full text-lg transition duration-300 ease-in-out transform hover:scale-105 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2">
            Ready to Start!
          </button>
        </Link>
      </motion.div>

      <motion.p
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5, delay: 0.8 }}
        className="mt-4 text-gray-600"
      >
        Click to get started on your journey to citizenship
      </motion.p>
    </div>
  );
};

export default HeroSection;