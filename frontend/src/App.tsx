import UploadForm from "./components/UploadForm";

function App() {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-blue-600 text-white p-4 text-center text-xl font-bold">
        Image Processor
      </header>
      <UploadForm />
    </div>
  );
}

export default App;
