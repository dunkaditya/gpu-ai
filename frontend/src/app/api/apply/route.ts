import { NextResponse } from "next/server";
import { Resend } from "resend";

const resend = process.env.RESEND_API_KEY
  ? new Resend(process.env.RESEND_API_KEY)
  : null;

const NOTIFY_EMAILS = [
  "aditya@novacorein.com",
  "ranbir@novacorein.com",
  "aryamaan@novacorein.com",
];

export async function POST(request: Request) {
  let formData: FormData;
  try {
    formData = await request.formData();
  } catch {
    return NextResponse.json(
      { error: "Invalid request body" },
      { status: 400 }
    );
  }

  const firstName = formData.get("firstName") as string;
  const lastName = formData.get("lastName") as string;
  const email = formData.get("email") as string;
  const phone = formData.get("phone") as string;
  const role = formData.get("role") as string;
  const linkedin = formData.get("linkedin") as string;
  const message = formData.get("message") as string;
  const resume = formData.get("resume") as File | null;

  if (!firstName || !lastName || !email || !role) {
    return NextResponse.json(
      { error: "Missing required fields" },
      { status: 400 }
    );
  }

  const emailBody = `New GPU.ai job application from ${firstName} ${lastName}

Role: ${role}
Email: ${email}
Phone: ${phone || "Not provided"}
LinkedIn: ${linkedin || "Not provided"}

Message:
${message || "No message provided"}`;

  if (resend) {
    try {
      const attachments: { filename: string; content: Buffer }[] = [];
      if (resume && resume.size > 0) {
        const buffer = Buffer.from(await resume.arrayBuffer());
        attachments.push({ filename: resume.name, content: buffer });
      }

      const { data, error } = await resend.emails.send({
        from:
          process.env.RESEND_FROM_EMAIL || "GPU.ai <noreply@novacorein.com>",
        to: NOTIFY_EMAILS,
        subject: `GPU.ai Application: ${role} — ${firstName} ${lastName}`,
        text: emailBody,
        attachments,
      });
      if (error) {
        console.error("Resend error:", error);
        return NextResponse.json(
          { error: error.message },
          { status: 500 }
        );
      }
      console.log("Email sent:", data);
    } catch (err) {
      console.error("Failed to send email:", err);
      return NextResponse.json(
        { error: "Failed to send email" },
        { status: 500 }
      );
    }
  } else {
    console.log("RESEND_API_KEY not set. Application logged:");
    console.log(emailBody);
    if (resume) console.log(`Resume attached: ${resume.name} (${resume.size} bytes)`);
  }

  return NextResponse.json({ success: true });
}
