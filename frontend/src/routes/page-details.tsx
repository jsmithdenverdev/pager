import { MapContainer, TileLayer, Marker, Popup } from "react-leaflet";

const pageDetails = `Loc: MOUNT BLUE SKY - 14000 HWY 5,MT. EVANS RD/MT. EVANS RD,39.5883,-105.6438,NOT FOUND,,07/05/2024 08:17:39,ALPINE//CCEN1//CCFN1,
[1] PSAP=JCEC--JEFFERSON WIRELESS VERIFYPD  VERIFY FD
[2] 1 MILE UP FROM SUMMIT LAKE TRHLHEAD
[3] SOMEONE COLLAPSED / UNCONCIOUS
[4] Multi-Agency Law Incident #: 07052024-0007020
[5] SOMEONE ACTIVATED SOS [Shared]
[6] RP IS 50 YARDS UP THE TRAIL FROM THE PT [Shared]
[7] >>> CONFIRMING ON THE ROAD OR BACKCOUNTRY [Shared]
[8] [Page] Problem changed from Cardiac/Resp Arrest/Death to Backcountry Rescue by Fire [Shared]
[9] Automatic Case Number(s) issued for Incident #[2024CCF-0001049], Jurisdiction: Clear Creek Fire. Case Number(s): 24-CCF-000895. requested by CCFN1. [Shared]
[10] Automatic Case Number(s) issued for Incident #[2024CCF-0001049], Jurisdiction: Clear Creek EMS. Case Number(s): 24-CCEMS-0961. requested by CCEN1. [Shared]
[11] ON THE TRAIL [Shared]
[12] [ProQA: Case Entry Complete] [Shared]
[13] >>AIRED [Shared]
[14] Automatic Case Number(s) issued for Incident #[2024CCF-0001049], Jurisdiction: Alpine Rescue Team. Case Number(s): 24-ALP-000049. requested by ALPINE. [Shared]`;
const Table = () => {
  return (
    <div className="overflow-x-auto">
      <table className="min-w-full table-auto border-collapse border border-gray-200">
        <thead className="hidden">
          <tr>
            <th className="border border-gray-300 px-4 py-2">Row Header</th>
            <th className="border border-gray-300 px-4 py-2">Column 1</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td className="border border-gray-300 px-4 py-2 font-bold">
              Response
            </td>
            <td className="border border-gray-300 px-4 py-2">24 total</td>
          </tr>
          <tr>
            <td className="border border-gray-300 px-4 py-2 font-bold">Log</td>
            <td className="border border-gray-300 px-4 py-2"></td>
          </tr>
          <tr>
            <td className="border border-gray-300 px-4 py-2 font-bold">
              Location (common name)
            </td>
            <td className="border border-gray-300 px-4 py-2">Summit Lake</td>
          </tr>
          <tr>
            <td className="border border-gray-300 px-4 py-2 font-bold">
              Units
            </td>
            <td className="border border-gray-300 px-4 py-2">ALPINE</td>
          </tr>
          <tr>
            <td className="border border-gray-300 px-4 py-2 font-bold">Date</td>
            <td className="border border-gray-300 px-4 py-2">
              Jul 5, 2024 at 08:18:25 MDT
            </td>
          </tr>
          <tr>
            <td className="border border-gray-300 px-4 py-2 font-bold">CAD</td>
            <td className="border border-gray-300 px-4 py-2">2024CCF-001049</td>
          </tr>
        </tbody>
      </table>
    </div>
  );
};

const buttons = [
  "Responding",
  "Team Veh.",
  "Cancelled",
  "Available",
  "601/605",
  "Med Suprt",
  "609",
  "609/Resp.",
  "608",
  "608/Resp.",
];

export default function PageDetails() {
  return (
    <>
      <section>
        <div className="cursor-pointer rounded-sm bg-blue-950 px-2 py-4 shadow hover:bg-blue-800">
          <div className="flex">
            <h1 className="text-2xl font-semibold uppercase text-gray-200">
              Backcountry Rescue
            </h1>
          </div>
        </div>
      </section>
      <section>
        <section className="space-y-2">
          <h2 className="text-xl font-semibold uppercase">Summit Lake</h2>
          <MapContainer
            center={[39.59894235, -105.64365396867784]}
            zoom={15}
            scrollWheelZoom={false}
            style={{ height: "260px" }}
          >
            <TileLayer
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors'
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />
            <Marker position={[51.505, -0.09]}>
              <Popup>
                A pretty CSS3 popup. <br /> Easily customizable.
              </Popup>
            </Marker>
          </MapContainer>
        </section>
      </section>

      <section className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
        {buttons.map((text, i) => (
          <button
            className="p-4 bg-slate-700 text-white uppercase hover:bg-slate-600"
            key={`button-${i}`}
            onClick={() => alert(text)}
          >
            {text}
          </button>
        ))}
      </section>
      <section>
        <Table />
      </section>
      <section>
        <h2 className="text-lg font-semibold">Details</h2>
        <pre className="max-w-screen-lg text-wrap p-2 bg-slate-800 text-white">{pageDetails}</pre>
      </section>
    </>
  );
}
