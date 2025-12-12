"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { LoginForm } from "@/components/auth/login-form";
import { RegisterForm } from "@/components/auth/register-form";

export default function AuthPage() {
  const [isLogin, setIsLogin] = useState(true);
  const router = useRouter();

  const handleAuthSuccess = () => {
    router.push("/");
  };

  const toggleMode = () => {
    setIsLogin(!isLogin);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        {isLogin
          ? (
              <LoginForm onToggleMode={toggleMode} onSuccess={handleAuthSuccess} />
            )
          : (
              <RegisterForm onToggleMode={toggleMode} onSuccess={handleAuthSuccess} />
            )}
      </div>
    </div>
  );
}
