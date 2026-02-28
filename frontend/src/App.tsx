import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import { AuthProvider } from "@/hooks/useAuth";
import { AuthGuard } from "@/components/auth/AuthGuard";
import { AppShell } from "@/components/layout/AppShell";
import LoginPage from "@/pages/LoginPage";
import AuthCallbackPage from "@/pages/AuthCallbackPage";
import DashboardPage from "@/pages/DashboardPage";
import TransactionsPage from "@/pages/TransactionsPage";
import CategoriesPage from "@/pages/CategoriesPage";
import BudgetsPage from "@/pages/BudgetsPage";
import RulesPage from "@/pages/RulesPage";
import WorkspaceSettingsPage from "@/pages/WorkspaceSettingsPage";
import BankCallbackPage from "@/pages/BankCallbackPage";
import InvitePage from "@/pages/InvitePage";
import NotFoundPage from "@/pages/NotFoundPage";

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          {/* Public routes */}
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/auth/:provider/callback"
            element={<AuthCallbackPage />}
          />

          {/* Bank auth callback (public, but redirects to settings) */}
          <Route path="/bank/callback" element={<BankCallbackPage />} />

          {/* Invite page (public, handles SSO bridge) */}
          <Route path="/invite/:code" element={<InvitePage />} />

          {/* Protected routes */}
          <Route
            element={
              <AuthGuard>
                <AppShell />
              </AuthGuard>
            }
          >
            <Route path="/dashboard" element={<DashboardPage />} />
            <Route path="/transactions" element={<TransactionsPage />} />
            <Route path="/categories" element={<CategoriesPage />} />
            <Route path="/budgets" element={<BudgetsPage />} />
            <Route path="/rules" element={<RulesPage />} />
            <Route path="/settings" element={<WorkspaceSettingsPage />} />
          </Route>

          {/* Redirects */}
          <Route path="/" element={<Navigate to="/dashboard" replace />} />
          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
