import rawConfig from "./devices.json";
import * as  React from "react";
import ReactDOM from "react-dom";

type Size = {
    width: number;
    height: number;
};

type Rect = Size & {
    x: number;
    y: number;
};

type DeviceConfiguration = {
    size: Size;
    ranges: {
        title: Rect;
        choice1: Rect;
        choice2: Rect;
        choice3: Rect;
    }
};

type BitmapImageForDevice = {
    title: ImageBitmap;
    choice1: ImageBitmap;
    choice2: ImageBitmap;
    choice3: ImageBitmap;
}

type BlobForDevice = {
    title: Blob;
    choice1: Blob;
    choice2: Blob;
    choice3: Blob;
}

type Choice = {
	choice: string;
	effect: string;
}

type Result = {
	ok: boolean;
	eventName: string;
	choices:   Choice[];
}

type DeviceConfigurations = {[key: string]: DeviceConfiguration};

const deviceConfigs: DeviceConfigurations = rawConfig;

function findConfig(width: number, height: number): DeviceConfiguration | undefined {
    const sorted = Object.entries(deviceConfigs).map(([title, config]) => {
        return {
            title,
            diff: Math.abs(config.size.height / config.size.width - height / width),
            config: config,
        }
    }).sort((a, b)=> a.diff - b.diff);

    if (!sorted.length || sorted[0].diff > 0.1) {
        return undefined;
    }

    return sorted[0].config;
}

const newTimer = (interval: number, handler: () => void) => {
    const identifier = setInterval(handler, interval);

    return () => {
        clearInterval(identifier);
    }
};


async function cropImage(image: ImageBitmapSource, rect: Rect): Promise<ImageBitmap> {
    return createImageBitmap(image, rect.x, rect.y, rect.width, rect.height);
}

function normRect(part: Rect, base: Size, target: Size): Rect {
    return {
        x: part.x / base.width * target.width,
        y: part.y / base.height * target.height,
        width: part.width / base.width * target.width,
        height: part.height / base.height * target.height,
    }
}

async function getJpegFromBitmap(img: ImageBitmap, parent?: HTMLElement) {
    const newCanvas = !parent.hasChildNodes();

    const canvas = newCanvas ? document.createElement('canvas') : parent.firstChild as HTMLCanvasElement;
    canvas.width = img.width;
    canvas.height = img.height;
    document.querySelector
    let ctx = canvas.getContext('bitmaprenderer');
    if(ctx) {
        ctx.transferFromImageBitmap(img);
    } else {
        canvas.getContext('2d').drawImage(img, 0, 0);
    }
    if (newCanvas) {
        parent.appendChild(canvas);
    }

    return await (new Promise<Blob>((resolve, reject) => {
        canvas.toBlob((blob) => {
            if (!blob) {
                reject("failed to encode the image as jpeg");
            }

            resolve(blob);
        }, 'image/jpeg', 0.5);
    }));
}

async function photoHandler(blob: ImageBitmap) {
    const all = await createImageBitmap(blob);

    const config = findConfig(all.width, all.height);

    if (!config) 
        return;
    
    const dom = {
        title: document.querySelector("#title"),
        choice1: document.querySelector("#choice1"),
        choice2: document.querySelector("#choice2"),
        choice3: document.querySelector("#choice3"),
    }

    const entries = await Promise.all(Object.entries(config.ranges).map(async ([key, rect]) => {
        const cropped = await cropImage(blob, normRect(rect, config.size, all))

        return [
            key,
            await getJpegFromBitmap(cropped, dom[key]),
        ]
    }));

    const images = Object.fromEntries(entries) as BlobForDevice;

    const formData = new FormData();
    
    for (const [name, image] of Object.entries(images)) {
        formData.append(name, image);
    }

    try {
        const res: Result = await fetch("/upload", {
            method: 'POST',
            body: formData,
        }).then(res => res.json());

        return res;
    } catch(e: any) {
        console.error(e);
    }
}

const entrypoint = async function() {
    const constraints: MediaStreamConstraints = {
        audio: false,
        video: true,
    };

    const stream = await ((navigator.mediaDevices as any).getDisplayMedia(constraints) as Promise<MediaStream>);
    const track = stream.getVideoTracks()[0];
    const imageCapture = new ImageCapture(track);

    const ResultTable: React.FC<{}> = function() {
        const [result, setResult] = React.useState<Result>(null);

        React.useEffect(() => {
            return newTimer(1000, () => {
                imageCapture.grabFrame()
                    .then(photoHandler)
                    .then((res) => {
                        setResult(res);
                    })
                    .catch(err => console.error(err))
            })
        }, []);

        return (<>
            <span>イベント名: {result?.eventName}</span>
            <table>
                <thead>
                    <tr>
                        <th>選択肢</th>
                        <th>結果</th>
                    </tr>
                </thead>
                <tbody>
                    {result?.choices.sort().map((c) => (
                        <tr key={c.choice}>
                            <td>{c.choice}</td>
                            <td>{c.effect}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </>)
    }

    ReactDOM.render(
        <ResultTable/>,
        document.querySelector("#table"),
    );

    const video = document.createElement("video")
    video.srcObject = stream;
    video.playsInline = true;
    video.muted = true;
    video.autoplay = true;
    
    document.querySelector("#video_wrapper")
        .appendChild(video);
}

entrypoint().catch((err) => {
    console.error(err);
})
