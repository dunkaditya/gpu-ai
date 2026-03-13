export const REGION_INFO: Record<string, { flag: string; label: string }> = {
  "us-east":      { flag: "\u{1F1FA}\u{1F1F8}", label: "US East" },
  "us-central":   { flag: "\u{1F1FA}\u{1F1F8}", label: "US Central" },
  "us-west":      { flag: "\u{1F1FA}\u{1F1F8}", label: "US West" },
  "ca-central":   { flag: "\u{1F1E8}\u{1F1E6}", label: "Canada" },
  "eu-west":      { flag: "\u{1F1EA}\u{1F1FA}", label: "EU West" },
  "eu-east":      { flag: "\u{1F1EA}\u{1F1FA}", label: "EU East" },
  "eu-north":     { flag: "\u{1F1EA}\u{1F1FA}", label: "EU North" },
  "ap-southeast": { flag: "\u{1F30F}", label: "Asia Pacific SE" },
  "ap-northeast": { flag: "\u{1F30F}", label: "Asia Pacific NE" },
  "ap-south":     { flag: "\u{1F1EE}\u{1F1F3}", label: "Asia South" },
  "me-central":   { flag: "\u{1F30D}", label: "Middle East" },
  "sa-east":      { flag: "\u{1F1E7}\u{1F1F7}", label: "South America" },
  "unknown":      { flag: "\u{1F310}", label: "All Regions" },
};

export function getRegionDisplay(region: string) {
  return REGION_INFO[region] ?? { flag: "\u{1F310}", label: region };
}
