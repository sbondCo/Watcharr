import { noAuthAxios } from "./api";

interface Plex {
  win: Window;
  headers: Record<string, string>;
  clientId: string;
}

interface PlexPin {
  id: number;
  code: string;
}

//
function uuidv4() {
  return "10000000-1000-4000-8000-100000000000".replace(/[018]/g, (c: any) =>
    (c ^ (crypto.getRandomValues(new Uint8Array(1))[0] & (15 >> (c / 4)))).toString(16)
  );
}

export function preparePlexAuth(): Plex {
  // No point making a store for this.
  let clientId = localStorage.getItem("plex-cid");
  if (!clientId) {
    if (crypto.randomUUID) {
      console.log("preparePlexAuth: Using randomUUID");
      clientId = crypto.randomUUID();
    } else {
      // Use this if randomUUID is unavailable (ex in unsecure context, this use case isn't as big a deelio)
      console.log("preparePlexAuth: Not using randomUUID");
      clientId = uuidv4();
    }
    localStorage.setItem("plex-cid", clientId);
  }

  const w = 600;
  const h = 800;
  // Fixes dual-screen position                           Most browsers       Firefox
  const dualScreenLeft = window.screenLeft != undefined ? window.screenLeft : window.screenX;
  const dualScreenTop = window.screenTop != undefined ? window.screenTop : window.screenY;
  const width = window.innerWidth
    ? window.innerWidth
    : document.documentElement.clientWidth
      ? document.documentElement.clientWidth
      : screen.width;
  const height = window.innerHeight
    ? window.innerHeight
    : document.documentElement.clientHeight
      ? document.documentElement.clientHeight
      : screen.height;
  const left = width / 2 - w / 2 + dualScreenLeft;
  const top = height / 2 - h / 2 + dualScreenTop;

  const win = window.open(
    "",
    "Continue to Watcharr With Plex",
    `scrollbars=yes, width=${w}, height=${h}, top=${top}, left=${left}`
  );
  if (win) {
    win.focus();
    return {
      win,
      // Don't really want to give all the possible headers,
      // trying to minimize it to what gets it working.
      headers: {
        "X-Plex-Product": "Watcharr",
        "X-Plex-Client-Identifier": clientId,
        "X-Plex-Version": "Plex OAuth",
        "X-Plex-Model": "Plex OAuth",
        "X-Plex-Platform": "Firefox",
        "X-Plex-Platform-Version": "123.0",
        "X-Plex-Device": "Windows",
        "X-Plex-Device-Name": "Watcharr"
      },
      clientId
    };
  }
  throw new Error("Failed to prepare popup!");
}

export async function doPlexLogin({ win, headers, clientId }: Plex): Promise<PlexPin> {
  const pin = await getPlexPin(headers);
  win.location.href = `https://app.plex.tv/auth/#!?clientID=${clientId}&code=${pin.code}&context=Watcharr&context[device][device]=${headers["X-Plex-Device"]}&context[device][deviceName]=${headers["X-Plex-Device-Name"]}&context[device][platform]=${headers["X-Plex-Platform"]}&context[device][platformVersion]=${headers["X-Plex-Platform-Version"]}&context[device][product]=${headers["X-Plex-Product"]}`;
  return pin;
}

export async function plexPinPoll(
  pin: PlexPin,
  { win, headers }: Plex,
  done: (err?: Error, token?: string) => void
) {
  const doPoll = async () => {
    try {
      console.debug("plexPinPoll");
      const r = await noAuthAxios.get(`https://plex.tv/api/v2/pins/${pin.id}`, {
        headers: { ...headers, code: pin.code }
      });
      if (r.data?.authToken) {
        done(undefined, r.data.authToken);
        win.close();
      } else if (win?.closed) {
        done(new Error("Plex popup closed before login completed"));
      } else {
        setTimeout(doPoll, 1000);
      }
    } catch (err) {
      console.error("plexPinPoll: Failed!", err);
      done(new Error("Pin poll failed"));
    }
  };
  setTimeout(doPoll, 1000);
}

async function getPlexPin(headers: Record<string, string>): Promise<PlexPin> {
  const r = await noAuthAxios.post(`https://plex.tv/api/v2/pins?strong=true`, undefined, {
    headers: headers
  });
  console.debug("getPlexPin:", r.data);
  return { id: r.data.id, code: r.data.code };
}
