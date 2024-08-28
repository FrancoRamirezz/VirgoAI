import Image from "next/image"
//import BronchoScoop from "ml-hack/public/Images/BronchoScoop.jpg"
// use this for the i
//<Image
//src={BronchoScoop}
//alt=""
//width={325}
//height={325}
//className="hidden md:block md:relative md:bottom-4 md:left-32 md:z-0"

const char = [
    { char: "" },
    { char: "Driven" },
    { char: "Night Owl" },
    {char:"Group Study"},
    {char: "Animal lover"},
    {char:"Pomodoro tech"},
    {char:"Pokemon music"},
    {char:"Documentary"},
    {char:"Podcast"},
    {char:"Rap music"}
  
  ]
  
  const Courses = () => {
    return (
       <main> 
      <section id="about">
        <div className="my-12 pb-12 md:pt-16 md:pb-48">
          <h1 className="text-center font-bold text-4xl">
            Welcome to the course
            <hr className="w-6 h-1 mx-auto my-4 bg-teal-500 border-0 rounded"></hr>
          </h1>
  
          <div className="flex flex-col space-y-10 items-stretch justify-center align-top md:space-x-10 md:space-y-0 md:p-4 md:flex-row md:text-left">
            <div className="md:w-1/2 ">
              <h1 className="text-center text-2xl font-bold mb-6 md:text-left">
                Get to know me!
              </h1>
              <p>
                Hello,my name is Franco and I'm looking for a new study buddy, I want them to be{" "}
                <span className="font-bold">{"night owl"}</span>,
                <span className="font-bold">{"music lover"}</span>,
                <span className="font-bold">{" driven"}</span> of course,
                must be a CPP student
              </p>
              <br />
              <br />
              <p>

                These are some characteristics that I can ask. Once you click all them then you can connect with others who have the same perfernces
              </p>
              <br />
              <p>
                College is diffcult already so{" "}
                <span className="font-bold text-teal-500">
                  let's study together
                </span>{" "}
                and that&#39; I'm looking  🙂
              </p>
            </div>
            <div className="text-center md:w-1/2 md:text-left">
              <h1 className="text-2xl font-bold mb-6">Study buddy characteristics</h1>
              <div className="flex flex-wrap flex-row justify-center z-10 md:justify-start">
                {char .map((item, idx) => {
                  return (
                    <p
                      key={idx}
                      className="bg-gray-200 px-4 py-2 mr-2 mt-2 text-gray-500 rounded font-semibold"
                    >
                      {item.char}
                    </p>
                  )
                })}
              </div>
            </div>
            
          
          </div>
        </div>
      </section>
      </main>
    )
  }
  
  export default Courses