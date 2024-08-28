'use client';

import React, { useState } from 'react';

const PaymentPage = () => {
  const [plan, setPlan] = useState('monthly');

  return (
    <div className="min-h-screen bg-gradient-to-b from-purple-100 to-blue-100 p-8">
      <header className="mb-12 text-center">
        <h1 className="text-4xl font-bold text-purple-600 mb-4">Choose Your Learning Plan</h1>
        <p className="text-xl text-gray-700">Unlock your English potential today!</p>
      </header>

      <main className="max-w-2xl mx-auto">
        <div className="bg-white rounded-lg shadow-md p-6 mb-8">
          <h2 className="text-2xl font-semibold text-center mb-4">Select Your Plan</h2>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <label className={`flex flex-col items-center justify-between rounded-md border-2 p-4 cursor-pointer ${plan === 'monthly' ? 'border-purple-500 bg-purple-50' : 'border-gray-200 hover:bg-gray-50'}`}>
              <input
                type="radio"
                name="plan"
                value="monthly"
                checked={plan === 'monthly'}
                onChange={() => setPlan('monthly')}
                className="sr-only"
              />
              <span className="text-xl font-semibold">Monthly Plan</span>
              <span className="text-3xl font-bold">$15.00</span>
              <span className="text-sm text-gray-500">Billed monthly</span>
            </label>
            <label className={`flex flex-col items-center justify-between rounded-md border-2 p-4 cursor-pointer ${plan === 'annual' ? 'border-purple-500 bg-purple-50' : 'border-gray-200 hover:bg-gray-50'}`}>
              <input
                type="radio"
                name="plan"
                value="annual"
                checked={plan === 'annual'}
                onChange={() => setPlan('annual')}
                className="sr-only"
              />
              <span className="text-xl font-semibold">Annual Plan</span>
              <span className="text-3xl font-bold">$199.99</span>
              <span className="text-sm text-gray-500">Billed annually (Save 17%)</span>
            </label>
          </div>
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-2xl font-semibold mb-4">Payment Details</h2>
          <form className="space-y-4">
            <div>
              <label htmlFor="cardName" className="block text-sm font-medium text-gray-700 mb-1">Name on Card</label>
              <input
                type="text"
                id="cardName"
                placeholder="John Doe"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
              />
            </div>
            <div>
              <label htmlFor="cardNumber" className="block text-sm font-medium text-gray-700 mb-1">Card Number</label>
              <input
                type="text"
                id="cardNumber"
                placeholder="1234 5678 9012 3456"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
              />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label htmlFor="expiry" className="block text-sm font-medium text-gray-700 mb-1">Expiry Date</label>
                <input
                  type="text"
                  id="expiry"
                  placeholder="MM/YY"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                />
              </div>
              <div>
                <label htmlFor="cvv" className="block text-sm font-medium text-gray-700 mb-1">CVV</label>
                <input
                  type="text"
                  id="cvv"
                  placeholder="123"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-purple-500"
                />
              </div>
            </div>
            <button
              type="submit"
              className="w-full bg-purple-600 hover:bg-purple-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:ring-2 focus:ring-purple-500 focus:ring-offset-2"
            >
              Pay {plan === 'monthly' ? '$19.99' : '$199.99'}
            </button>
          </form>
        </div>

        <div className="mt-8 text-center text-sm text-gray-600">
          <p>By clicking "Pay", you agree to our Terms of Service and Privacy Policy.</p>
          <p className="mt-2">Need help? Contact our support team at support@example.com</p>
        </div>
      </main>
    </div>
  );
};

export default PaymentPage;

