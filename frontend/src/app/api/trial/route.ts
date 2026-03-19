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
  let body: Record<string, string>;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json(
      { error: "Invalid request body" },
      { status: 400 }
    );
  }

  const {
    firstName,
    lastName,
    email,
    phone,
    company,
    jobTitle,
    gpuModel,
    workload,
    monthlySpend,
    timeline,
  } = body;

  if (!firstName || !lastName || !email || !phone || !company || !jobTitle) {
    return NextResponse.json(
      { error: "Missing required fields" },
      { status: 400 }
    );
  }

  const emailBody = `New GPU.ai free trial request from ${firstName} ${lastName}

Company: ${company}
Job Title: ${jobTitle}
Email: ${email}
Phone: ${phone}

GPU Model Interest: ${gpuModel || "Not specified"}
Primary Workload: ${workload || "Not specified"}
Estimated Monthly Spend: ${monthlySpend || "Not specified"}
Timeline: ${timeline || "Not specified"}`;

  if (resend) {
    try {
      const { data, error } = await resend.emails.send({
        from:
          process.env.RESEND_FROM_EMAIL || "GPU.ai <noreply@novacorein.com>",
        to: NOTIFY_EMAILS,
        subject: `GPU.ai Trial Request: ${company} — ${firstName} ${lastName}`,
        text: emailBody,
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
    console.log("RESEND_API_KEY not set. Submission logged:");
    console.log(emailBody);
  }

  return NextResponse.json({ success: true });
}
