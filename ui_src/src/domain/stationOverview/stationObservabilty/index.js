// Copyright 2021-2022 The Memphis Authors
// Licensed under the MIT License (the "License");
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// This license limiting reselling the software itself "AS IS".

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

import './style.scss';

import React, { useContext, useEffect, useRef } from 'react';
import Lottie from 'lottie-react';

import consumePoision from '../../../assets/lotties/consume_poision.json';
import consumeEmpty from '../../../assets/lotties/consume_empty.json';
import produceEmpty from '../../../assets/lotties/produce_empty.json';
import produce from '../../../assets/lotties/produce-many.json';
import consumer from '../../../assets/lotties/consume.json';
import ProduceConsumList from './ProduceConsumList';
import { StationStoreContext } from '..';
import Messages from './messages';

const StationObservabilty = () => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);

    return (
        <div className="station-observabilty-container">
            <ProduceConsumList producer={true} />
            <div className="thunnel-from-sub">
                {stationState?.stationSocketData?.connected_producers?.length === 0 && <Lottie animationData={produceEmpty} loop={true} />}
                {stationState?.stationSocketData?.connected_producers?.length > 0 && <Lottie animationData={produce} loop={true} />}
            </div>
            <Messages />
            <div className="thunnel-to-pub">
                {stationState?.stationSocketData?.connected_cgs?.length === 0 && <Lottie animationData={consumeEmpty} loop={true} />}
                {stationState?.stationSocketData?.connected_cgs?.length > 0 && stationState?.stationSocketData?.poison_messages?.length > 0 && (
                    <Lottie animationData={consumePoision} loop={true} />
                )}
                {stationState?.stationSocketData?.connected_cgs?.length > 0 && stationState?.stationSocketData?.poison_messages?.length === 0 && (
                    <Lottie animationData={consumer} loop={true} />
                )}
            </div>
            <ProduceConsumList producer={false} />
        </div>
    );
};

export default StationObservabilty;
