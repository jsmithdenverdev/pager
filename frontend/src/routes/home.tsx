export default function Home() {
  return (
    <div className="space-y-4">
      <section className="space-y-4">
        <div className="cursor-pointer rounded-sm bg-blue-950 p-4 shadow hover:bg-blue-800">
          <div className="flex">
            <h2 className="text-xl font-semibold uppercase text-gray-200">
              Heat / Cold Exposure
            </h2>
            <p className="text-md ml-auto font-semibold uppercase text-gray-200">
              11:46 - 10/16/24
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-200">600 silver city road</p>
          </div>
        </div>
        <div className="cursor-pointer rounded-sm bg-blue-950 p-4 shadow hover:bg-blue-800">
          <div className="flex">
            <h2 className="text-xl font-semibold uppercase text-gray-200">
              ART - FINAL STAND DOWN
            </h2>
            <p className="text-md ml-auto font-semibold uppercase text-gray-200">
              15:08 - 10/13/24
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-200">19000 Hwy 103</p>
          </div>
        </div>
        <div className="cursor-pointer rounded-sm bg-blue-950 p-4 shadow hover:bg-blue-800">
          <div className="flex">
            <h2 className="text-xl font-semibold uppercase text-gray-200">
              ART - Respond Emergent to Echo Lake Lodge
            </h2>
            <p className="text-md ml-auto font-semibold uppercase text-gray-200">
              9:30 - 10/13/24
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-200">19000 Hwy 103</p>
          </div>
        </div>
        <div className="cursor-pointer rounded-sm bg-blue-950 p-4 shadow hover:bg-blue-800">
          <div className="flex">
            <h2 className="text-xl font-semibold uppercase text-gray-200">
              Backcountry Rescue
            </h2>
            <p className="text-md ml-auto font-semibold uppercase text-gray-200">
              9:24 - 10/13/24
            </p>
          </div>
          <div>
            <p className="text-sm text-gray-200">19000 Hwy 103</p>
          </div>
        </div>
      </section>
    </div>
  );
}
