import { ref } from "vue";
import { Browser } from "@wailsio/runtime";

export interface ReleaseAsset {
  name: string;
  size: number;
  browser_download_url: string;
}

export interface ReleaseInfo {
  tagName: string;
  name: string;
  notes: string;
  publishedAt: string;
  url: string;
  assets: ReleaseAsset[];
}
export const CURRENT_VERSION = "v0.2.1";
const checking = ref(false);
const hasUpdate = ref(false);
const latestRelease = ref<ReleaseInfo | null>(null);
const isModalOpen = ref(false);
const feedbackMessage = ref("");
const feedbackType = ref<"success" | "info" | "error" | "">("");

export function useUpdaterStore() {
  function compareVersions(v1: string, v2: string): number {
    const s1 = v1.replace(/^v/, "").split(".").map((n) => parseInt(n, 10) || 0);
    const s2 = v2.replace(/^v/, "").split(".").map((n) => parseInt(n, 10) || 0);
    for (let i = 0; i < Math.max(s1.length, s2.length); i++) {
      const num1 = s1[i] || 0;
      const num2 = s2[i] || 0;
      if (num1 > num2) return 1;
      if (num1 < num2) return -1;
    }
    return 0;
  }

  async function checkForUpdates(manual = true) {
    checking.value = true;
    feedbackMessage.value = "";
    feedbackType.value = "";

    try {
      const resp = await fetch("https://api.github.com/repos/lianchengwu/vectordb-gui/releases/latest", {
        headers: {
          Accept: "application/vnd.github.v3+json",
        },
      });

      if (!resp.ok) {
        if (resp.status === 404) {
          if (manual) {
            feedbackMessage.value = `当前已是最新版本 (${CURRENT_VERSION})`;
            feedbackType.value = "info";
            isModalOpen.value = true;
          }
          return;
        }
        throw new Error(`GitHub API HTTP ${resp.status}`);
      }

      const data: unknown = await resp.json();
      if (!data || typeof data !== "object") {
        throw new Error("Invalid GitHub response");
      }

      const record = data as Record<string, unknown>;
      const tagName = typeof record.tag_name === "string" ? record.tag_name : "";
      const name = typeof record.name === "string" ? record.name : tagName;
      const notes = typeof record.body === "string" ? record.body : "";
      const publishedAt = typeof record.published_at === "string" ? record.published_at : "";
      const url = typeof record.html_url === "string" ? record.html_url : "";

      const rawAssets = Array.isArray(record.assets) ? record.assets : [];
      const assets: ReleaseAsset[] = [];
      for (const item of rawAssets) {
        if (item && typeof item === "object") {
          const a = item as Record<string, unknown>;
          assets.push({
            name: typeof a.name === "string" ? a.name : "",
            size: typeof a.size === "number" ? a.size : 0,
            browser_download_url: typeof a.browser_download_url === "string" ? a.browser_download_url : "",
          });
        }
      }

      latestRelease.value = {
        tagName,
        name,
        notes,
        publishedAt,
        url,
        assets,
      };

      if (compareVersions(tagName, CURRENT_VERSION) > 0) {
        hasUpdate.value = true;
        feedbackType.value = "success";
        isModalOpen.value = true;
      } else {
        hasUpdate.value = false;
        if (manual) {
          feedbackMessage.value = `当前已是最新版本 (${CURRENT_VERSION})`;
          feedbackType.value = "info";
          isModalOpen.value = true;
        }
      }
    } catch (e: unknown) {
      if (manual) {
        feedbackMessage.value = `检查更新失败: ${e instanceof Error ? e.message : String(e)}`;
        feedbackType.value = "error";
        isModalOpen.value = true;
      }
    } finally {
      checking.value = false;
    }
  }

  function openReleaseUrl(targetUrl?: string) {
    const link = targetUrl || latestRelease.value?.url || "https://github.com/lianchengwu/vectordb-gui/releases";
    try {
      Browser.OpenURL(link);
    } catch {
      window.open(link, "_blank");
    }
  }

  function closeModal() {
    isModalOpen.value = false;
  }

  return {
    CURRENT_VERSION,
    checking,
    hasUpdate,
    latestRelease,
    isModalOpen,
    feedbackMessage,
    feedbackType,
    checkForUpdates,
    openReleaseUrl,
    closeModal,
  };
}
