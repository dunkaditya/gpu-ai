import { CLI_COMMANDS } from "@/lib/constants";
import { Container } from "@/components/ui";

export function CLIDemo() {
  return (
    <section className="py-24">
      <Container>
        {/* Section heading */}
        <div className="text-center mb-12">
          <h2 className="text-3xl md:text-4xl font-bold text-white">
            Powerful CLI, zero friction
          </h2>
          <p className="mt-4 text-text-muted text-lg">
            Launch, monitor, and connect to GPU instances with a single tool.
          </p>
        </div>

        {/* Terminal window */}
        <div className="mx-auto max-w-[800px] rounded-xl border border-white/10 bg-[#0a0a0a] overflow-hidden">
          {/* Chrome bar */}
          <div className="flex items-center gap-2 border-b border-white/10 px-4 py-3">
            <div className="flex gap-1.5">
              <span className="h-3 w-3 rounded-full bg-[#ff5f56]" />
              <span className="h-3 w-3 rounded-full bg-[#ffbd2e]" />
              <span className="h-3 w-3 rounded-full bg-[#27c93f]" />
            </div>
            <span className="ml-2 text-xs text-gray-500">terminal</span>
          </div>

          {/* Code area */}
          <div className="p-6 font-mono text-sm space-y-4">
            {CLI_COMMANDS.map((cmd, index) => (
              <div key={index}>
                <div>
                  <span className="text-green-400">{cmd.prompt}</span>{" "}
                  <span className="text-white">{cmd.command}</span>
                </div>
                {cmd.output.map((line, lineIndex) => (
                  <div key={lineIndex} className="text-text-muted pl-4">
                    {line}
                  </div>
                ))}
              </div>
            ))}
          </div>
        </div>
      </Container>
    </section>
  );
}
