#!/usr/bin/env python3
import json, sys
from collections import defaultdict

if __name__ == '__main__':
    import argparse
    parser = argparse.ArgumentParser()
    parser.add_argument("trace_in")
    parser.add_argument("parsed_trace")
    args = parser.parse_args()

    events = []
    min_ts = 0
    current_color = 0
    color_names = [
        "thread_state_uninterruptible",
        "thread_state_iowait",
        "thread_state_running",
        "thread_state_runnable",
        "thread_state_unknown"]

    elapseds = {}
    n_fr = n_fd = 0

    names = {}
    with open(args.trace_in, 'r') as f:
        for line in f:
            if "file log configured" in line:
                continue
            d = json.loads(line)
            if d['name'] == 'Parlink' and d['ph'] == 'B':
                min_ts = d['ts']
                
            d['ts'] -= min_ts
            d.pop('level', None)
        
            #find the FD that started earlier and finished latest
            if d['name'] == 'FaceDetection':
                if d["pid"] not in elapseds:
                    elapseds[d["pid"]] = {"min_fd": sys.maxsize, "max_fd": 0, "min_fr": sys.maxsize, "max_fr": 0}

                elapseds[d["pid"]]["min_fd"] = min(d['ts'], elapseds[d["pid"]]["min_fd"])
                elapseds[d["pid"]]["max_fd"] = max(d['ts'], elapseds[d["pid"]]["max_fd"])
                n_fd += 1
                #d['true_name'] = 'FD'

            #rename for better show
            if d['name'] == 'face_recognition_unitwork_part2' or d['name'] == 'FaceRecognition':
                if d["pid"] not in elapseds:
                    elapseds[d["pid"]] = {"min_fd": sys.maxsize, "max_fd": 0, "min_fr": sys.maxsize, "max_fr": 0}
                d['name'] = 'FaceRecognition'
                #d['true_name'] = 'FR'
                #find the FR that started earlier and finished latest
                elapseds[d["pid"]]["min_fr"] = min(d['ts'], elapseds[d["pid"]]["min_fr"])
                elapseds[d["pid"]]["max_fr"] = max(d['ts'], elapseds[d["pid"]]["max_fr"])
                n_fr += 1

            events.append(d)

            if d['name'] not in names:
                print(f"{d['name']}, {color_names[current_color]}")
                names[d['name']] = color_names[current_color]
                current_color += 1
                current_color %= len(color_names)
            if current_color > len(color_names):
                print("not enough color schemes")
    for e in events:
        if e['name'] == "Encode":
            color_name = names[e['name']]
            e['cname'] = color_name
    parsed_event = {}
    parsed_event["traceEvents"] = events
    with open(args.parsed_trace, 'w') as f:
        json.dump(parsed_event, f, indent=4)


    print("Face Detection: \n")
    fd_elapseds = []
    fr_elapseds = []
    for pid, data in elapseds.items():
        print(data)
        #print("FD, {}, FR, {}".format(data["max_fd"]-data["min_fd"], data["max_fr"]-data["min_fr"] ))
        fd_elapseds.append((data["max_fd"]-data["min_fd"])/1000000)
        fr_elapseds.append((data["max_fr"]-data["min_fr"])/1000000)

    print("face detection times:", fd_elapseds)
    print("face recognition times:", fr_elapseds)
    print("\n\n")
    print(f"face detection min/max: {min(fd_elapseds)} , {max(fd_elapseds)}")
    print(f"face recognition min/max: {min(fr_elapseds)} , {max(fr_elapseds)}")
