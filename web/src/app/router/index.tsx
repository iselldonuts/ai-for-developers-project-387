import { Route, Routes } from "react-router-dom";
import { HomePage } from "../../pages/home-page";
import { OwnerPage } from "../../pages/owner-page";

export function AppRouter() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/owner" element={<OwnerPage />} />
    </Routes>
  );
}
