const Sidebar = () =>{

return(
<div className="w-64 h-full bg-white shadow-md">
      <div className="p-6">
        <h1 className="text-2xl font-semibold">ShadCN Dashboard</h1>
      </div>
      <nav className="mt-10">
        <a className="block py-2.5 px-4 rounded transition duration-200 hover:bg-gray-200" href="#">
          Dashboard
        </a>
        <a className="block py-2.5 px-4 rounded transition duration-200 hover:bg-gray-200" href="#">
          Settings
        </a>
        <a className="block py-2.5 px-4 rounded transition duration-200 hover:bg-gray-200" href="#">
          Profile
        </a>
      </nav>
    </div>


)


}
export default Sidebar;