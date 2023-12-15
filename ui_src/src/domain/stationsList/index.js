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

import React, { useEffect, useContext, useState, useRef } from 'react';
import { Virtuoso } from 'react-virtuoso';
import { useGetAllowedActions } from '../../services/genericServices';
import { ReactComponent as DeleteWrapperIcon } from '../../assets/images/deleteWrapperIcon.svg';
import StationsInstructions from '../../components/stationsInstructions';
import { ReactComponent as StationIcon } from '../../assets/images/stationIcon.svg';
import CreateStationForm from '../../components/createStationForm';
import { stationFilterArray, isCloud } from '../../services/valueConvertor';
import DeleteItemsModal from '../../components/deleteItemsModal';
import stationsIcon from '../../assets/images/stationIcon.svg';
import { ApiEndpoints } from '../../const/apiEndpoints';
import StationBoxOverview from './stationBoxOverview';
import { httpRequest } from '../../services/http';
import Button from '../../components/button';
import Filter from '../../components/filter';
import Loader from '../../components/loader';
import LearnMore from '../../components/learnMore';
import { Context } from '../../hooks/store';
import Modal from '../../components/modal';
import CloudModal from '../../components/cloudModal';
import { FaArrowCircleUp } from 'react-icons/fa';
import RefreshButton from '../../components/refreshButton';

const StationsList = () => {
    const [state, dispatch] = useContext(Context);
    const [modalIsOpen, modalFlip] = useState(false);
    const [modalDeleteIsOpen, modalDeleteFlip] = useState(false);
    const [isLoading, setisLoading] = useState(true);
    const [deleteLoader, setDeleteLoader] = useState(false);
    const [creatingProsessd, setCreatingProsessd] = useState(false);
    const [isCheck, setIsCheck] = useState([]);
    const [isCheckAll, setIsCheckAll] = useState(false);
    const [openCloudModal, setOpenCloudModal] = useState(false);
    const createStationRef = useRef(null);
    const getAllowedActions = useGetAllowedActions();

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'stations' });
        getAllStations();
        return () => {
            dispatch({ type: 'SET_STATION_LIST', payload: [] });
            dispatch({ type: 'SET_SCHEMA_FILTERED_LIST', payload: [] });
        };
    }, []);

    const getAllStations = async () => {
        setisLoading(true);
        try {
            const res = await httpRequest('GET', `${ApiEndpoints.GET_STATIONS}`);
            res.stations.sort((a, b) => new Date(b.station.created_at) - new Date(a.station.created_at));
            dispatch({ type: 'SET_STATION_LIST', payload: res.stations });
            dispatch({ type: 'SET_SCHEMA_FILTERED_LIST', payload: res.stations });
            setTimeout(() => {
                setisLoading(false);
            }, 500);
        } catch (err) {
            setisLoading(false);
            return;
        }
    };

    const renderStationsOverview = () => {
        if (state?.stationList?.length > 0) {
            if (state.stationFilteredList?.length === 0) {
                return <StationsInstructions header="No stations found" des="Please try to search again" image={stationsIcon} />;
            }
            if (state?.stationList?.length <= 2) {
                return (
                    <div>
                        {state.stationFilteredList?.map((station) => (
                            <StationBoxOverview
                                key={station?.station?.id}
                                isCheck={isCheck?.includes(station?.station?.name)}
                                handleCheckedClick={handleCheckedClick}
                                station={station}
                            />
                        ))}
                        <div className="stations-placeholder add-more">
                            <Button
                                className="modal-btn"
                                width="230px"
                                height="42px"
                                placeholder={
                                    isCloud() && !state?.allowedActions?.can_create_stations ? (
                                        <span className="create-new">
                                            <label>Add another station</label>
                                            <FaArrowCircleUp className="lock-feature-icon" />
                                        </span>
                                    ) : (
                                        <span className="create-new">Add another station</span>
                                    )
                                }
                                colorType="white"
                                radiusType="circle"
                                backgroundColorType="purple"
                                fontSize="16px"
                                fontWeight="bold"
                                onClick={() => (!isCloud() || state?.allowedActions?.can_create_stations ? modalFlip(true) : setOpenCloudModal(true))}
                            />
                        </div>
                    </div>
                );
            }
            return (
                <Virtuoso
                    data={state?.stationFilteredList}
                    overscan={100}
                    itemContent={(index, station) => (
                        <StationBoxOverview
                            key={station?.station?.id}
                            isCheck={isCheck?.includes(station?.station?.name)}
                            handleCheckedClick={handleCheckedClick}
                            station={station}
                        />
                    )}
                />
            );
        }
        return (
            <StationsInstructions
                upgrade={!state?.allowedActions?.can_create_stations}
                header="You don’t have any station yet"
                button="Create a new station"
                image={stationsIcon}
                newStation={() => modalFlip(true)}
            />
        );
    };

    const onCheckedAll = (e) => {
        setIsCheckAll(!isCheckAll);
        setIsCheck(state.stationFilteredList.map((li) => li.station?.name));
        if (isCheckAll) {
            setIsCheck([]);
        }
    };

    const handleCheckedClick = (e) => {
        const { id, checked } = e.target;
        setIsCheck([...isCheck, id]);
        if (!checked) {
            setIsCheck(isCheck.filter((item) => item !== id));
        }
        if (isCheck.length === 1 && !checked) {
            setIsCheckAll(false);
        }
    };

    const handleDeleteSelected = async () => {
        setDeleteLoader(true);
        try {
            const data = await httpRequest('DELETE', ApiEndpoints.REMOVE_STATION, {
                station_names: isCheck
            });
            if (data) {
                dispatch({ type: 'SET_STATION_LIST', payload: stationFilterArray(state?.stationFilteredList, isCheck) });
                setIsCheck([]);
                setIsCheckAll(false);
                setDeleteLoader(false);
                modalDeleteFlip(false);
            }
        } catch (error) {
            setDeleteLoader(false);
        } finally {
            getAllowedActions();
        }
    };

    return (
        <div className="stations-details-container">
            <div className="stations-details-header">
                <div className="header-wraper">
                    <div className="main-header-wrapper">
                        <label className="main-header-h1">
                            Stations <label className="length-list">{state?.stationFilteredList?.length > 0 && `(${state?.stationFilteredList?.length})`}</label>
                        </label>
                        <span className="memphis-label">
                            A station is a distributed storage for messages. More&nbsp;
                            <a className="learn-more" href=" https://docs.memphis.dev/memphis/memphis/key-concepts/station" target="_blank">
                                here.
                            </a>
                        </span>
                    </div>
                    <div className="right-side">
                        <RefreshButton onClick={() => getAllStations()} isLoading={isLoading} />
                        <Button
                            height="34px"
                            placeholder={`Delete selected (${isCheck?.length})`}
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            fontSize="12px"
                            fontWeight="600"
                            aria-haspopup="true"
                            boxShadowStyle="float"
                            disabled={isCheck?.length === 0}
                            onClick={() => modalDeleteFlip(true)}
                        />
                        <Button
                            width="131px"
                            height="34px"
                            placeholder={isCheckAll ? 'Unselect all' : 'Select all'}
                            colorType="black"
                            radiusType="circle"
                            backgroundColorType="white"
                            fontSize="12px"
                            fontWeight="600"
                            aria-haspopup="true"
                            boxShadowStyle="float"
                            disabled={state?.stationFilteredList?.length === 0}
                            onClick={() => onCheckedAll()}
                        />
                        <Filter filterComponent="stations" height="34px" />
                        <Button
                            width="170px"
                            height="34px"
                            placeholder={
                                isCloud() && !state?.allowedActions?.can_create_stations ? (
                                    <span className="create-new">
                                        <label>Create a new station</label>
                                        <FaArrowCircleUp className="lock-feature-icon" />
                                    </span>
                                ) : (
                                    <span className="create-new">Create a new station</span>
                                )
                            }
                            colorType="white"
                            radiusType="circle"
                            backgroundColorType="purple"
                            fontSize="12px"
                            boxShadowStyle="float"
                            fontWeight="600"
                            aria-haspopup="true"
                            onClick={() => (!isCloud() || state?.allowedActions?.can_create_stations ? modalFlip(true) : setOpenCloudModal(true))}
                        />
                    </div>
                </div>
            </div>
            {isLoading && (
                <div className="loader-uploading">
                    <Loader />
                </div>
            )}
            {!isLoading && <div className="stations-content">{renderStationsOverview()}</div>}
            <div>
                <Modal
                    header={
                        <div className="modal-header">
                            <div className="header-img-container">
                                <StationIcon alt="stationIcon" />
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
                    open={modalIsOpen}
                    isLoading={creatingProsessd}
                >
                    <CreateStationForm createStationFormRef={createStationRef} setLoading={(e) => setCreatingProsessd(e)} />
                </Modal>
            </div>
            <Modal
                header={<DeleteWrapperIcon alt="deleteWrapperIcon" />}
                width="520px"
                height="240px"
                displayButtons={false}
                clickOutside={() => modalDeleteFlip(false)}
                open={modalDeleteIsOpen}
            >
                <DeleteItemsModal
                    title="Are you sure you want to delete the selected stations?"
                    desc="Deleting these stations means they will be permanently deleted."
                    buttontxt="I understand, delete the selected stations"
                    handleDeleteSelected={handleDeleteSelected}
                    loader={deleteLoader}
                />
            </Modal>
            <CloudModal type="upgrade" open={openCloudModal} handleClose={() => setOpenCloudModal(false)} />
        </div>
    );
};

export default StationsList;
