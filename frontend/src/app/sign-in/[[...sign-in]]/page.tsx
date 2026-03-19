import { SignIn } from "@clerk/nextjs";
import { ChipLogo } from "@/components/ui/ChipLogo";

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-bg relative">
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[600px] rounded-full bg-purple/10 blur-3xl" />
      </div>
      <div className="relative z-10 flex flex-col items-center gap-4">
        <a href="/" className="flex items-center gap-1">
          <ChipLogo size={42} />
          <span className="font-sans text-[26px] font-bold tracking-[-0.5px]">
            <span className="text-white">gpu</span>
            <span className="gradient-text">.ai</span>
          </span>
        </a>
        <SignIn forceRedirectUrl="/cloud" />
      </div>
    </div>
  );
}
