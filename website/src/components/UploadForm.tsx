import axios from "axios";
import { useEffect, useRef, useState } from "react";
import { useDropzone } from "react-dropzone";

interface Action {
  name: string;
  params?: Record<string, string>;
}

interface UploadedImage {
  id: string;
  filename: string;
  path: string;
  status: string;
  preview?: string; // локальный preview
  action?: {
    name: string;
    params: Record<string, string>;
  };
}

export default function UploadForm() {
  const [files, setFiles] = useState<File[]>([]);
  const [action, setAction] = useState<Action>({
    name: "resize",
    params: { width: "200", height: "200" },
  });
  const [uploadedImages, setUploadedImages] = useState<UploadedImage[]>([]);
  const uploadedImagesRef = useRef<UploadedImage[]>([]);
  const [watermarkText, setWatermarkText] = useState("");

  const onDrop = (acceptedFiles: File[]) => setFiles(acceptedFiles);
  const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop });

  // держим актуальный ref
  useEffect(() => {
    uploadedImagesRef.current = uploadedImages;
  }, [uploadedImages]);

  // создаем локальные preview для выбранных файлов
  const previews = files.map((file) => URL.createObjectURL(file));

  const handleUpload = async () => {
    if (!files.length) return;

    const formData = new FormData();
    formData.append("image", files[0]);

    const params: Record<string, string> =
      action.name === "watermark"
        ? { text: watermarkText }
        : {
            width: action.params?.width || "200",
            height: action.params?.height || "200",
          };

    formData.append("actions", JSON.stringify({ action: action.name, params }));

    try {
      const { data } = await axios.post<{ result: UploadedImage }>(
        "http://localhost:8080/api/upload",
        formData,
        { headers: { "Content-Type": "multipart/form-data" } }
      );

      // создаём локальный preview и сразу показываем pending
      setUploadedImages((prev) => [
        ...prev,
        {
          ...data.result,
          status: "pending",
          preview: previews[0],
          action: { ...action, params },
        },
      ]);

      // очищаем выбранные файлы и input
      setFiles([]);
    } catch (err) {
      console.error(err);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await axios.delete(`http://localhost:8080/api/image/${id}`);
      setUploadedImages((prev) => {
        const imgToDelete = prev.find((img) => img.id === id);
        if (imgToDelete?.preview) URL.revokeObjectURL(imgToDelete.preview);
        return prev.filter((img) => img.id !== id);
      });
    } catch (err) {
      console.error(err);
    }
  };

  // при размонтировании очищаем все локальные URL
  useEffect(() => {
    return () => {
      uploadedImages.forEach(
        (img) => img.preview && URL.revokeObjectURL(img.preview)
      );
      previews.forEach((url) => URL.revokeObjectURL(url));
    };
  }, []);

  // polling для обновления статуса изображений
  useEffect(() => {
    const interval = setInterval(async () => {
      const pendingImages = uploadedImagesRef.current.filter(
        (img) => img.status !== "processed"
      );
      if (!pendingImages.length) return;

      try {
        const updatedImages = await Promise.all(
          pendingImages.map(async (img) => {
            const { data } = await axios.get<{ result: UploadedImage }>(
              `http://localhost:8080/api/image/${img.id}/meta`
            );
            return { ...img, ...data.result };
          })
        );

        setUploadedImages((prev) =>
          prev.map((img) => {
            const updated = updatedImages.find((u) => u.id === img.id);
            return updated ? updated : img;
          })
        );
      } catch (err) {
        console.error(err);
      }
    }, 2000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="p-6 max-w-3xl mx-auto">
      <div
        {...getRootProps()}
        className={`flex flex-col items-center justify-center border-2 border-dashed p-6 h-86 text-center rounded-lg cursor-pointer ${
          isDragActive ? "border-blue-500" : "border-gray-300"
        }`}
      >
        <input {...getInputProps()} />
        {files.length > 0 ? (
          <div className="flex flex-col items-center gap-2">
            <img
              src={previews[0]}
              alt={files[0].name}
              className="w-32 h-32 object-cover rounded"
            />
            <p className="text-sm">{files[0].name}</p>
          </div>
        ) : isDragActive ? (
          <p>Drop the image here ...</p>
        ) : (
          <p>Drag & drop an image here, or click to select</p>
        )}
      </div>

      <div className="mt-4 flex flex-col md:flex-row items-center gap-4">
        <select
          className="border rounded p-2"
          value={action.name}
          onChange={(e) =>
            setAction({
              name: e.target.value,
              params: { width: "200", height: "200" },
            })
          }
        >
          <option value="resize">Resize</option>
          <option value="thumbnail">Thumbnail</option>
          <option value="watermark">Watermark</option>
        </select>

        {(action.name === "resize" || action.name === "thumbnail") && (
          <div className="flex gap-2">
            <input
              type="number"
              min={1}
              placeholder="Width"
              value={action.params?.width}
              onChange={(e) =>
                setAction((prev) => ({
                  ...prev,
                  params: { ...prev.params, width: e.target.value },
                }))
              }
              className="border rounded p-2 w-20"
            />
            <input
              type="number"
              min={1}
              placeholder="Height"
              value={action.params?.height}
              onChange={(e) =>
                setAction((prev) => ({
                  ...prev,
                  params: { ...prev.params, height: e.target.value },
                }))
              }
              className="border rounded p-2 w-20"
            />
          </div>
        )}

        {action.name === "watermark" && (
          <input
            type="text"
            placeholder="Watermark text"
            value={watermarkText}
            onChange={(e) => setWatermarkText(e.target.value)}
            className="border rounded p-2 flex-1"
          />
        )}

        <button
          onClick={handleUpload}
          className="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600"
        >
          Upload
        </button>
      </div>

      <div className="mt-6 grid grid-cols-2 md:grid-cols-3 gap-4">
        {uploadedImages.map((img) => (
          <div key={img.id} className="border p-2 rounded">
            <img
              key={`${img.id}-${img.status}`}
              src={`http://localhost:8080/api/image/${img.id}?t=${Date.now()}`}
              alt={img.filename}
              className="w-full h-40"
            />
            <p className="text-sm mt-1">{img.filename}</p>
            <p className="text-xs text-gray-500">{img.status}</p>
            <button
              onClick={() => handleDelete(img.id)}
              className="mt-2 bg-red-500 text-white px-2 py-1 rounded hover:bg-red-600 text-sm"
            >
              Delete
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
