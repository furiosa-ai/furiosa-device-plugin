import sys
from furiosa.models.vision import SSDMobileNet
from furiosa.runtime.sync import create_runner

# make sure that test images exist in right place
image = ["license_free_cat.jpg"]

mobilenet = SSDMobileNet()
with create_runner(mobilenet.model_source()) as runner:
    inputs, contexts = mobilenet.preprocess(image)
    outputs = runner.run(inputs)
    converted = mobilenet.postprocess(outputs, contexts)
    for inner_list in converted:
        for detection in inner_list:
            print(f"Label: {detection.label}, Score: {detection.score}")
sys.exit()
