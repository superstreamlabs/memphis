// Copyright 2022-2023 The Memphis.dev Authors
// Licensed under the Memphis Business Source License 1.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// Changed License: [Apache License, Version 2.0 (https://www.apache.org/licenses/LICENSE-2.0), as published by the Apache Foundation.
//
// https://github.com/memphisdev/memphis/blob/master/LICENSE
//
// Additional Use Grant: You may make use of the Licensed Work (i) only as part of your own product or service, provided it is not a message broker or a message queue product or service; and (ii) provided that you do not use, provide, distribute, or make available the Licensed Work as a Service.
// A "Service" is a commercial offering, product, hosted, or managed service, that allows third parties (other than your own employees and contractors acting on your behalf) to access and/or use the Licensed Work or a substantial set of the features or functionality of the Licensed Work to third parties as a software-as-a-service, platform-as-a-service, infrastructure-as-a-service or other similar services that compete with Licensor products or services.

import './style.scss';

import React, { useContext } from 'react';
import Lottie from 'lottie-react';

import consumePoison from 'assets/lotties/consume_poison.json';
import consumeEmpty from 'assets/lotties/consume_empty.json';
import produceEmpty from 'assets/lotties/produce_empty.json';
import produce from 'assets/lotties/produce-many.json';
import consumer from 'assets/lotties/consume.json';
import ProduceConsumList from './ProduceConsumList';
import { StationStoreContext } from '..';
import Messages from './messages';

const StationObservabilty = ({ referredFunction, loading }) => {
    const [stationState, stationDispatch] = useContext(StationStoreContext);

    return (
        <div className="station-observabilty-container">
            <ProduceConsumList producer={true} />
            <div className="thunnel-from-sub">
                {stationState?.stationMetaData?.is_native ? (
                    <>
                        {stationState?.stationSocketData?.connected_producers?.length === 0 &&
                            stationState?.stationSocketData?.source_connectors?.filter((connector) => connector?.is_active)?.length === 0 && (
                                <Lottie animationData={produceEmpty} loop={true} />
                            )}
                        {(stationState?.stationSocketData?.connected_producers?.length > 0 ||
                            stationState?.stationSocketData?.source_connectors?.filter((connector) => connector?.is_active)?.length > 0) && (
                            <Lottie animationData={produce} loop={true} />
                        )}
                    </>
                ) : (
                    <Lottie animationData={produceEmpty} loop={true} />
                )}
            </div>

            <Messages referredFunction={referredFunction} loading={loading} />

            <div className="thunnel-to-pub">
                {stationState?.stationMetaData?.is_native ? (
                    <>
                        {stationState?.stationSocketData?.connected_cgs?.length === 0 &&
                            stationState?.stationSocketData?.sink_connectors?.filter((connector) => connector?.is_active)?.length === 0 && (
                                <Lottie animationData={consumeEmpty} loop={true} />
                            )}
                        {(stationState?.stationSocketData?.connected_cgs?.length > 0 ||
                            stationState?.stationSocketData?.sink_connectors?.filter((connector) => connector?.is_active)?.length > 0) &&
                            stationState?.stationSocketData?.poison_messages?.length > 0 && <Lottie animationData={consumePoison} loop={true} />}
                        {(stationState?.stationSocketData?.connected_cgs?.length > 0 ||
                            stationState?.stationSocketData?.sink_connectors?.filter((connector) => connector?.is_active)?.length > 0) &&
                            stationState?.stationSocketData?.poison_messages?.length === 0 && <Lottie animationData={consumer} loop={true} />}
                    </>
                ) : (
                    <Lottie animationData={consumeEmpty} loop={true} />
                )}
            </div>
            <ProduceConsumList producer={false} />
        </div>
    );
};

export default StationObservabilty;
