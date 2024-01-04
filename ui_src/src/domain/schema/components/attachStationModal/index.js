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

import React, { useEffect, useRef, useState } from 'react';

import CheckboxComponent from 'components/checkBox';
import { ReactComponent as AttachedPlaceholderIcon } from 'assets/images/attachedPlaceholder.svg';
import { ReactComponent as StationsActiveIcon } from 'assets/images/stationsIconActive.svg';
import { ApiEndpoints } from 'const/apiEndpoints';
import { httpRequest } from 'services/http';
import { useGetAllowedActions } from 'services/genericServices';
import Button from 'components/button';
import OverflowTip from 'components/tooltip/overflowtip';
import Modal from 'components/modal';
import LearnMore from 'components/learnMore';
import CreateStationForm from 'components/createStationForm';
import { isCloud } from 'services/valueConvertor';

function AttachStationModal({ close, handleAttachedStations, attachedStations, schemaName, update }) {
    const createStationRef = useRef(null);
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [allStations, setAllStations] = useState([]);
    const [attachLoader, setAttachLoader] = useState(false);
    const [indeterminate, setIndeterminate] = useState(false);
    const [creatingProsessd, setCreatingProsessd] = useState(false);
    const [open, modalFlip] = useState(false);

    const getAllowedActions = useGetAllowedActions();
    const onCheckedAll = (e) => {
        if (!update && attachedStations?.length > 0) {
            setIndeterminate(!indeterminate);
            if (indeterminate) {
                setIsCheck([]);
            } else {
                allStations?.map((li) => {
                    if (attachedStations?.includes(li.name)) return;
                    else setIsCheck(...li.name);
                });
            }
        } else {
            setIsCheckAll(!isCheckAll);
            setIndeterminate(false);
            if (isCheckAll) {
                setIsCheck([]);
            } else {
                setIsCheck(allStations?.map((li) => li.name));
            }
        }
    };

    const handleCheckedClick = (id, checked) => {
        if (!update && attachedStations?.includes(id)) return;
        let checkedList = [];
        if (!checked) {
            setIsCheck(isCheck?.filter((item) => item !== id));
            checkedList = isCheck?.filter((item) => item !== id);
        }
        if (checked) {
            checkedList = [...isCheck, id];
            setIsCheck(checkedList);
        }
        setIsCheckAll(checkedList?.length === allStations?.length);
        setIndeterminate(!!checkedList?.length && checkedList?.length < allStations?.length);
    };

    const getAllStations = async () => {
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_ALL_STATIONS}`);
            let native_station = res.filter((station) => station.is_native);
            if (update) {
                let attachedStation = native_station.filter((station) => {
                    return attachedStations?.includes(station.name);
                });
                setAllStations(attachedStation);
            } else {
                setAllStations(native_station);
            }
        } catch (err) {
            return;
        }
    };

    useEffect(() => {
        getAllStations();

        return () => {
            setAllStations([]);
            setIsCheck([]);
        };
    }, []);

    const attachToStation = async () => {
        setAttachLoader(true);
        try {
            const data = await httpRequest('POST', ApiEndpoints.USE_SCHEMA, { station_names: isCheck, schema_name: schemaName });
            if (data) {
                !update && handleAttachedStations(isCheck);
                setAttachLoader(false);
                close();
            }
        } catch (error) {
            setAttachLoader(false);
            close();
        } finally {
            isCloud() && getAllowedActions();
        }
    };

    return (
        <div className="attach-station-content">
            <p className="title">{update ? 'Enforce the new version' : 'Enforce a schema on a station'}</p>
            <p className="desc">{update ? 'Which stations should be updated' : 'Enforcing a scheme on a station will force the producers to comply with it'}</p>
            <div className="stations-list">
                {allStations?.length > 0 ? (
                    <div className="header">
                        <CheckboxComponent
                            disabled={!update && attachedStations?.length === allStations?.length}
                            indeterminate={indeterminate}
                            checked={isCheckAll}
                            id={'selectAll'}
                            onChange={onCheckedAll}
                            name={'selectAll'}
                        />
                        <p className="ovel-label">Station Name</p>
                    </div>
                ) : (
                    <div className="placeholder">
                        <AttachedPlaceholderIcon alt="attachedPlaceholder" />
                        <p>No stations yet</p>
                        <Button
                            className="modal-btn"
                            width="160px"
                            height="34px"
                            placeholder={'Create a new station'}
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontWeight="600"
                            aria-haspopup="true"
                            boxShadowStyle="float"
                            onClick={() => modalFlip(true)}
                        />
                    </div>
                )}
                {allStations?.length > 0 && (
                    <div className="station-wraper">
                        {allStations?.map((station, index) => {
                            return (
                                <div
                                    key={station.name}
                                    className="station-row"
                                    onClick={() => handleCheckedClick(station.name, isCheck?.includes(station.name) ? false : true)}
                                >
                                    <CheckboxComponent
                                        disabled={!update && attachedStations?.includes(station.name)}
                                        checked={isCheck?.includes(station.name) || (!update && attachedStations?.includes(station.name))}
                                        id={station.name}
                                        onChange={(e) => handleCheckedClick(e.target.id, e.target.checked)}
                                        name={station.name}
                                    />
                                    <OverflowTip className="ovel-label" text={station.name}>
                                        {station.name}
                                    </OverflowTip>
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>
            <div className="buttons" style={{ justifyContent: allStations?.length > 0 ? 'space-between' : 'flex-end' }}>
                <>
                    <Button
                        width="150px"
                        height="34px"
                        placeholder="Cancel"
                        colorType="black"
                        radiusType="circle"
                        backgroundColorType="white"
                        border="gray-light"
                        fontSize="12px"
                        fontFamily="InterSemiBold"
                        onClick={() => close()}
                    />
                    {allStations?.length > 0 && (
                        <Button
                            width="150px"
                            height="34px"
                            placeholder="Enforce Selected"
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            fontFamily="InterSemiBold"
                            loading={attachLoader}
                            disabled={isCheck?.length === 0}
                            onClick={() => attachToStation()}
                        />
                    )}
                </>
            </div>
            <Modal
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <StationsActiveIcon alt="stationsIconActive" className="headerImage" />
                        </div>
                        <p>Create a new station</p>
                        <label>
                            A station is a distributed unit that stores the produced data{' '}
                            <LearnMore url="https://docs.memphis.dev/memphis/memphis-broker/concepts/station" />
                        </label>
                    </div>
                }
                height="58vh"
                width="1020px"
                rBtnText="Create"
                lBtnText="Cancel"
                lBtnClick={() => {
                    modalFlip(false);
                }}
                rBtnClick={() => {
                    createStationRef.current();
                }}
                clickOutside={() => modalFlip(false)}
                open={open}
                isLoading={creatingProsessd}
            >
                <CreateStationForm createStationFormRef={createStationRef} setLoading={(e) => setCreatingProsessd(e)} />
            </Modal>
        </div>
    );
}

export default AttachStationModal;
