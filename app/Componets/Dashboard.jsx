'use client'
import React, { useState, useEffect } from 'react';

const StudentDashboard = () => {
  const initialStudents = [
    { id: "001", name: "Alice Johnson", className: "US History", grade: "A", studentId: "ST12345", email: "alice@example.com", lastLogin: "2024-08-01" },
    { id: "002", name: "Bob Smith", className: "Civics", grade: "B", studentId: "ST12346", email: "bob@example.com", lastLogin: "2024-08-02" },
    { id: "003", name: "Charlie Brown", className: "English Language", grade: "A-", studentId: "ST12347", email: "charlie@example.com", lastLogin: "2024-08-03" },
    { id: "004", name: "Diana Ross", className: "US Government", grade: "B+", studentId: "ST12348", email: "diana@example.com", lastLogin: "2024-08-04" },
    { id: "005", name: "Ethan Hunt", className: "US History", grade: "C", studentId: "ST12349", email: "ethan@example.com", lastLogin: "2024-08-05" },
    { id: "006", name: "Fiona Gallagher", className: "Civics", grade: "A", studentId: "ST12350", email: "fiona@example.com", lastLogin: "2024-08-06" },
    { id: "007", name: "George Michael", className: "English Language", grade: "B-", studentId: "ST12351", email: "george@example.com", lastLogin: "2024-08-07" },
    { id: "008", name: "Hannah Montana", className: "US Government", grade: "A+", studentId: "ST12352", email: "hannah@example.com", lastLogin: "2024-08-08" },
  ];

  const [students, setStudents] = useState(initialStudents);
  const [sortConfig, setSortConfig] = useState({ key: null, direction: 'ascending' });
  const [searchTerm, setSearchTerm] = useState('');
  const [expandedRow, setExpandedRow] = useState(null);
  const [editingCell, setEditingCell] = useState({ rowId: null, field: null });
  const [currentPage, setCurrentPage] = useState(1);
  const [studentsPerPage] = useState(5);

  useEffect(() => {
    setCurrentPage(1);
  }, [searchTerm]);

  const sortData = (key) => {
    let direction = 'ascending';
    if (sortConfig.key === key && sortConfig.direction === 'ascending') {
      direction = 'descending';
    }
    setSortConfig({ key, direction });

    setStudents([...students].sort((a, b) => {
      if (a[key] < b[key]) return direction === 'ascending' ? -1 : 1;
      if (a[key] > b[key]) return direction === 'ascending' ? 1 : -1;
      return 0;
    }));
  };

  const filteredStudents = students.filter(student =>
    Object.values(student).some(value => 
      value.toString().toLowerCase().includes(searchTerm.toLowerCase())
    )
  );

  const indexOfLastStudent = currentPage * studentsPerPage;
  const indexOfFirstStudent = indexOfLastStudent - studentsPerPage;
  const currentStudents = filteredStudents.slice(indexOfFirstStudent, indexOfLastStudent);

  const paginate = (pageNumber) => setCurrentPage(pageNumber);

  const handleEdit = (rowId, field, value) => {
    setStudents(students.map(student => 
      student.id === rowId ? { ...student, [field]: value } : student
    ));
    setEditingCell({ rowId: null, field: null });
  };

  const getGradeColor = (grade) => {
    if (grade.startsWith('A')) return 'text-green-600';
    if (grade.startsWith('B')) return 'text-blue-600';
    if (grade.startsWith('C')) return 'text-yellow-600';
    return 'text-red-600';
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-purple-100 to-blue-100 p-8">
      <div className="max-w-6xl mx-auto bg-white shadow-lg rounded-lg overflow-hidden">
        <div className="p-6">
          <h2 className="text-2xl font-bold text-purple-700 mb-4">Student Dashboard</h2>
          <div className="flex justify-between items-center mb-4">
            <div className="relative">
              <input
                type="text"
                placeholder="Search students..."
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                className="pl-10 pr-4 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-purple-600"
              />
              <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5 text-gray-400 absolute left-3 top-3" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M8 4a4 4 0 100 8 4 4 0 000-8zM2 8a6 6 0 1110.89 3.476l4.817 4.817a1 1 0 01-1.414 1.414l-4.816-4.816A6 6 0 012 8z" clipRule="evenodd" />
              </svg>
            </div>
            <button className="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 focus:outline-none focus:ring-2 focus:ring-purple-600 focus:ring-opacity-50">
              Export Data
            </button>
          </div>
          <div className="overflow-x-auto">
            <table className="min-w-full">
              <thead className="bg-gray-50">
                <tr>
                  {['Student ID', 'Name', 'Class Name', 'Grade'].map((header, index) => (
                    <th 
                      key={index}
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                      onClick={() => sortData(header.toLowerCase().replace(' ', ''))}
                    >
                      {header}
                      <span className="ml-2">↕</span>
                    </th>
                  ))}
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {currentStudents.map((student) => (
                  <React.Fragment key={student.id}>
                    <tr className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">{student.studentId}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                        {editingCell.rowId === student.id && editingCell.field === 'name' ? (
                          <input
                            type="text"
                            value={student.name}
                            onChange={(e) => handleEdit(student.id, 'name', e.target.value)}
                            onBlur={() => setEditingCell({ rowId: null, field: null })}
                            className="border-b border-gray-300 focus:outline-none focus:border-blue-500"
                            autoFocus
                          />
                        ) : (
                          <span onClick={() => setEditingCell({ rowId: student.id, field: 'name' })}>{student.name}</span>
                        )}
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">{student.className}</td>
                      <td className={`px-6 py-4 whitespace-nowrap text-sm font-medium ${getGradeColor(student.grade)}`}>{student.grade}</td>
                      <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                        <button 
                          onClick={() => setExpandedRow(expandedRow === student.id ? null : student.id)}
                          className="text-purple-600 hover:text-purple-900"
                        >
                          {expandedRow === student.id ? 'Hide Details' : 'Show Details'}
                        </button>
                      </td>
                    </tr>
                    {expandedRow === student.id && (
                      <tr>
                        <td colSpan="5" className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 bg-gray-50">
                          <div>Email: {student.email}</div>
                          <div>Last Login: {student.lastLogin}</div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                ))}
              </tbody>
            </table>
          </div>
          <div className="mt-4 flex justify-between items-center">
            <div>
              Showing {indexOfFirstStudent + 1}-{Math.min(indexOfLastStudent, filteredStudents.length)} of {filteredStudents.length}
            </div>
            <div>
              {Array.from({ length: Math.ceil(filteredStudents.length / studentsPerPage) }, (_, i) => (
                <button
                  key={i}
                  onClick={() => paginate(i + 1)}
                  className={`mx-1 px-3 py-1 rounded ${currentPage === i + 1 ? 'bg-purple-600 text-white' : 'bg-gray-200'}`}
                >
                  {i + 1}
                </button>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StudentDashboard;