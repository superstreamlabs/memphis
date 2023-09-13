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

import React, { useEffect, useContext, useState, useRef } from 'react';
import { AccountCircleRounded } from '@material-ui/icons';

import { LOCAL_STORAGE_USER_PASS_BASED_AUTH } from '../../const/localStorageConsts';
import { isCloud, parsingDate } from '../../services/valueConvertor';
import addUserIcon from '../../assets/images/addUserIcon.svg';
import { ReactComponent as AddUserIcon } from '../../assets/images/addUserIcon.svg';
import deleteWrapperIcon from '../../assets/images/deleteWrapperIcon.svg';
import { ReactComponent as DeleteWrapperIcon } from '../../assets/images/deleteWrapperIcon.svg';
import mailIcon from '../../assets/images/mailIcon.svg';
import { ReactComponent as MailIcon } from '../../assets/images/mailIcon.svg';
import deleteIcon from '../../assets/images/deleteIcon.svg';
import { ReactComponent as DeleteIcon } from '../../assets/images/deleteIcon.svg';
import searchIcon from '../../assets/images/searchIcon.svg';
import { ReactComponent as SearchIcon } from '../../assets/images/searchIcon.svg';
import SegmentButton from '../../components/segmentButton';
import { ApiEndpoints } from '../../const/apiEndpoints';
import SearchInput from '../../components/searchInput';
import ActiveBadge from '../../components/activeBadge';
import CreateUserDetails from './createUserDetails';
import { httpRequest } from '../../services/http';
import Loader from '../../components/loader';
import Button from '../../components/button';
import { Context } from '../../hooks/store';
import Modal from '../../components/modal';
import Table from '../../components/table';
import DeleteItemsModal from '../../components/deleteItemsModal';

function Users() {
    const [state, dispatch] = useContext(Context);
    const [userList, setUsersList] = useState([]);
    const [copyOfUserList, setCopyOfUserList] = useState([]);
    const [addUserModalIsOpen, addUserModalFlip] = useState(false);
    const [userDetailsModal, setUserDetailsModal] = useState(false);
    const [removeUserModalOpen, setRemoveUserModalOpen] = useState(false);
    const [userToRemove, setuserToRemove] = useState('');
    const [userDeletedLoader, setUserDeletedLoader] = useState(false);
    const createUserRef = useRef(null);
    const [searchInput, setSearchInput] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const [tableType, setTableType] = useState('Management (0)');
    const [resendEmailLoader, setResendEmailLoader] = useState(false);
    const [createUserLoader, setCreateUserLoader] = useState(false);
    const [userToResend, setuserToResend] = useState('');

    useEffect(() => {
        dispatch({ type: 'SET_ROUTE', payload: 'users' });
        getAllUsers();
    }, [dispatch]);

    const getAllUsers = async () => {
        try {
            setIsLoading(true);
            const data = await httpRequest('GET', ApiEndpoints.GET_ALL_USERS);
            if (data) {
                data.management_users.sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
                data.application_users.sort((a, b) => new Date(a.created_at) - new Date(b.created_at));
                setTableType(`Management (${data?.management_users?.length || 0})`);
                setUsersList(data);
                setCopyOfUserList(data);
            }
        } catch (error) {}
        setIsLoading(false);
    };

    useEffect(() => {
        if (searchInput.length > 1) {
            let copy = copyOfUserList;
            const results = {
                management_users: copy.management_users.filter(
                    (userData) =>
                        userData?.username?.toLowerCase()?.includes(searchInput.toLowerCase()) || userData?.user_type?.toLowerCase()?.includes(searchInput.toLowerCase())
                ),
                application_users: copy.application_users.filter(
                    (userData) =>
                        userData?.username?.toLowerCase()?.includes(searchInput.toLowerCase()) || userData?.user_type?.toLowerCase()?.includes(searchInput.toLowerCase())
                )
            };
            if (tableType.includes('Management')) {
                setTableType(`Management (${results?.management_users?.length || 0})`);
            }

            if (tableType.includes('Client')) {
                setTableType(`Client (${results?.application_users?.length || 0})`);
            }
            setCopyOfUserList(results);
        } else {
            if (tableType.includes('Management')) {
                setTableType(`Management (${userList?.management_users?.length || 0})`);
            }

            if (tableType.includes('Client')) {
                setTableType(`Client (${userList?.application_users?.length || 0})`);
            }
            setCopyOfUserList(userList);
        }
    }, [searchInput]);

    const handleSearch = (e) => {
        setSearchInput(e.target.value);
    };

    const handleAddUser = (userData) => {
        setUsersList((prevUserData) => {
            const updatedUserData = { ...prevUserData };

            if (userData.user_type === 'management') {
                updatedUserData.management_users = [...updatedUserData.management_users, userData];
                setTableType(`Management (${updatedUserData?.management_users?.length || 0})`);
            }

            if (userData.user_type === 'application') {
                updatedUserData.application_users = [...updatedUserData.application_users, userData];
                setTableType(`Client (${updatedUserData?.application_users?.length || 0})`);
            }
            return updatedUserData;
        });

        setCopyOfUserList((prevUserData) => {
            const updatedUserData = { ...prevUserData };

            if (userData.user_type === 'management') {
                updatedUserData.management_users = [...updatedUserData.management_users, userData];
            }

            if (userData.user_type === 'application') {
                updatedUserData.application_users = [...updatedUserData.application_users, userData];
            }

            return updatedUserData;
        });

        addUserModalFlip(false);
        setCreateUserLoader(false);
        if (userData.user_type === 'application' && localStorage.getItem(LOCAL_STORAGE_USER_PASS_BASED_AUTH) === 'false') {
            setUserDetailsModal(true);
        }
    };

    const getAvatarSrc = (avatarId) => {
        return require(`../../assets/images/bots/avatar${avatarId}.svg`);
    };

    const handleRemoveUser = async (name, type) => {
        setUsersList((prevUserData) => {
            const updatedUserData = { ...prevUserData };
            updatedUserData[type] = updatedUserData[type].filter((user) => user.username !== name);
            if (type === 'management_users') {
                setTableType(`Management (${updatedUserData[type]?.length || 0})`);
            } else {
                setTableType(`Client (${updatedUserData[type]?.length || 0})`);
            }
            return updatedUserData;
        });

        setCopyOfUserList((prevUserData) => {
            const updatedUserData = { ...prevUserData };
            updatedUserData[type] = updatedUserData[type].filter((user) => user.username !== name);
            return updatedUserData;
        });
        setUserDeletedLoader(false);
        setRemoveUserModalOpen(false);
    };

    const removeUser = async (user) => {
        setUserDeletedLoader(true);
        try {
            await httpRequest(user.revoke ? 'POST' : 'DELETE', user.revoke ? ApiEndpoints.REVOKED_INVITATION : ApiEndpoints.REMOVE_USER, {
                username: user.username
            });
            handleRemoveUser(user.username, user.type === 'management' ? 'management_users' : 'application_users');
        } catch (error) {
            setUserDeletedLoader(false);
            setRemoveUserModalOpen(false);
        }
    };

    const deleteUser = (username, user_type) => {
        setuserToRemove({ username: username, type: user_type });
        setRemoveUserModalOpen(true);
    };

    const revokeUser = (username, user_type) => {
        setuserToRemove({ username: username, type: user_type, revoke: true });
        setRemoveUserModalOpen(true);
    };

    const resendEmail = async (username) => {
        setuserToResend(username);
        setResendEmailLoader(true);
        try {
            await httpRequest('POST', ApiEndpoints.RESEND_INVITATION, {
                username: username
            });
            setTimeout(() => {
                setResendEmailLoader(false);
            }, 1000);
        } catch (error) {
            setResendEmailLoader(false);
        }
        setuserToResend('');
    };

    const clientColumns = [
        {
            title: 'Username',
            dataIndex: 'username',
            key: 'username',
            render: (text) => (
                <div className="user-name">
                    <div className="user-avatar">
                        <AccountCircleRounded />
                    </div>
                    <p>{text}</p>
                </div>
            )
        },
        {
            title: 'Owner',
            key: 'owner',
            dataIndex: 'owner',
            render: (owner) => (
                <div className="owner">
                    <p>{owner || '-'}</p>
                </div>
            )
        },
        {
            title: 'Description',
            key: 'description',
            dataIndex: 'description',
            render: (description) => (
                <div className="created-column">
                    <p>{description || '-'}</p>
                </div>
            )
        },
        {
            title: 'Creation date',
            key: 'created_at',
            dataIndex: 'created_at',
            render: (created_at) => (
                <div className="created-column">
                    <p>{parsingDate(created_at)}</p>
                </div>
            )
        },
        {
            title: 'Action',
            dataIndex: 'action',
            key: 'action',
            render: (_, record) => (
                <div className="user-action">
                    <Button
                        width="115px"
                        height="30px"
                        placeholder={
                            <div className="action-button">
                                <DeleteIcon className="delete-icon" alt="deleteIcon" />
                                Delete user
                            </div>
                        }
                        colorType="red"
                        radiusType="circle"
                        border="gray-light"
                        backgroundColorType={'white'}
                        fontSize="12px"
                        fontFamily="InterMedium"
                        onClick={() => {
                            deleteUser(record.username, record.user_type);
                        }}
                    />
                </div>
            )
        }
    ];

    const managmentColumns = [
        {
            title: 'Username',
            dataIndex: 'username',
            key: 'username',
            render: (text, record) => (
                <div className="user-name">
                    <div className="user-avatar">
                        <img src={getAvatarSrc(record.avatar_id)} width={25} height={25} alt="avatar" />
                    </div>
                    <p>{text}</p>
                    {record.user_type === 'root' && <ActiveBadge active={false} content={isCloud() ? 'Owner' : 'Root'} />}
                </div>
            )
        },
        {
            title: 'Status',
            dataIndex: 'pending',
            key: 'pending',
            render: (pending) => (
                <div className="status">
                    <ActiveBadge active={!pending} content={pending ? 'Pending' : 'Active'} />
                </div>
            )
        },
        {
            title: 'Full name',
            key: 'full_name',
            dataIndex: 'full_name',
            render: (full_name) => (
                <div className="full-name">
                    <p>{full_name || '-'}</p>
                </div>
            )
        },
        {
            title: 'Created by',
            key: 'created_by',
            dataIndex: 'owner',
            render: (owner) => (
                <div className="created_by">
                    <p>{owner || '-'}</p>
                </div>
            )
        },
        {
            title: 'Team',
            key: 'team',
            dataIndex: 'team',
            render: (team) => (
                <div className="team">
                    <p>{team || '-'}</p>
                </div>
            )
        },
        {
            title: 'Position',
            key: 'position',
            dataIndex: 'position',
            render: (position) => (
                <div className="position">
                    <p>{position || '-'}</p>
                </div>
            )
        },
        {
            title: 'Creation date',
            key: 'created_at',
            dataIndex: 'created_at',
            render: (created_at) => (
                <div className="created-column">
                    <p>{parsingDate(created_at)}</p>
                </div>
            )
        },
        {
            title: 'Action',
            dataIndex: 'action',
            key: 'action',
            render: (_, record) =>
                record.user_type !== 'root' && (
                    <div className="user-action">
                        {isCloud() && record.pending ? (
                            <>
                                <Button
                                    width="125px"
                                    height="30px"
                                    placeholder={
                                        <div className="action-button">
                                            <MailIcon className="action-img-btn" alt="mailIcon" />
                                            Resend email
                                        </div>
                                    }
                                    colorType="black"
                                    radiusType="circle"
                                    border="gray-light"
                                    backgroundColorType={'white'}
                                    fontSize="12px"
                                    fontFamily="InterMedium"
                                    isLoading={record.username === userToResend && resendEmailLoader}
                                    onClick={() => {
                                        resendEmail(record.username);
                                    }}
                                />
                                <Button
                                    width="95px"
                                    height="30px"
                                    placeholder={
                                        <div className="action-button">
                                            <DeleteIcon className="action-img-btn" alt="deleteIcon" />
                                            Revoke
                                        </div>
                                    }
                                    colorType="red"
                                    radiusType="circle"
                                    border="gray-light"
                                    backgroundColorType={'white'}
                                    fontSize="12px"
                                    fontFamily="InterMedium"
                                    isLoading={record.username === userToRemove.username && userDeletedLoader}
                                    onClick={() => {
                                        revokeUser(record.username, record.user_type);
                                    }}
                                />
                            </>
                        ) : (
                            <Button
                                width="115px"
                                height="30px"
                                placeholder={
                                    <div className="action-button">
                                        <DeleteIcon className="action-img-btn" alt="deleteIcon" />
                                        Delete user
                                    </div>
                                }
                                colorType="red"
                                radiusType="circle"
                                border="gray-light"
                                backgroundColorType={'white'}
                                fontSize="12px"
                                fontFamily="InterMedium"
                                isLoading={record.username === userToRemove.username && userDeletedLoader}
                                onClick={() => {
                                    deleteUser(record.username, record.user_type);
                                }}
                            />
                        )}
                    </div>
                )
        }
    ];

    const changeTableView = (e) => {
        setTableType(e);
    };

    const tableHeader = () => {
        return (
            <div className="table-header">
                <p>User type:</p>
                <SegmentButton
                    value={tableType}
                    size="middle"
                    options={[`Management (${copyOfUserList?.management_users?.length || 0})`, `Client (${copyOfUserList?.application_users?.length || 0})`]}
                    onChange={(e) => changeTableView(e)}
                />
            </div>
        );
    };
    return (
        <div className="users-container">
            <div className="header-wraper">
                <div className="main-header-wrapper">
                    <label className="main-header-h1">Users</label>
                    <span className="memphis-label">For client authentication, choose "Client". For management only, choose "Management".</span>
                </div>
                <div className="add-search-user">
                    <SearchInput
                        placeholder="Search here"
                        colorType="navy"
                        backgroundColorType="gray-dark"
                        width="288px"
                        height="34px"
                        borderRadiusType="circle"
                        borderColorType="none"
                        boxShadowsType="none"
                        iconComponent={<SearchIcon alt="searchIcon" />}
                        onChange={handleSearch}
                        value={searchInput}
                    />
                    <Button
                        className="modal-btn"
                        width="160px"
                        height="34px"
                        placeholder={'Add new user'}
                        colorType="white"
                        radiusType="circle"
                        backgroundColorType="purple"
                        fontSize="12px"
                        fontWeight="600"
                        boxShadowStyle="float"
                        aria-haspopup="true"
                        onClick={() => addUserModalFlip(true)}
                    />
                </div>
            </div>
            <div className="users-list-container">
                {isLoading && (
                    <div className="loader-uploading">
                        <Loader />
                    </div>
                )}
                {!isLoading && (
                    <Table
                        className="users-table"
                        tableRowClassname="user-row"
                        title={tableHeader}
                        columns={tableType.includes('Management') ? managmentColumns : clientColumns}
                        data={tableType.includes('Management') ? copyOfUserList?.management_users : copyOfUserList?.application_users}
                    />
                )}
            </div>
            <Modal
                header={
                    <div className="modal-header">
                        <div className="header-img-container">
                            <AddUserIcon className="headerImage" alt="addUserIcon" />
                        </div>
                        <p>Add a new user</p>
                        <label>Enter user details to get started</label>
                    </div>
                }
                width="450px"
                rBtnText="Create"
                lBtnText="Cancel"
                lBtnClick={() => {
                    addUserModalFlip(false);
                    setCreateUserLoader(false);
                }}
                clickOutside={() => {
                    setCreateUserLoader(false);
                    addUserModalFlip(false);
                }}
                rBtnClick={() => {
                    setCreateUserLoader(true);
                    createUserRef.current();
                }}
                isLoading={createUserLoader}
                open={addUserModalIsOpen}
            >
                <CreateUserDetails
                    createUserRef={createUserRef}
                    userList={userList}
                    closeModal={(userData) => handleAddUser(userData)}
                    handleLoader={(e) => setCreateUserLoader(e)}
                />
            </Modal>
            <Modal
                header="User connection details"
                height="220px"
                rBtnText="Close"
                clickOutside={() => setUserDetailsModal(false)}
                rBtnClick={() => {
                    setUserDetailsModal(false);
                }}
                open={userDetailsModal}
            >
                <div className="user-details-modal">
                    <p className="userName">
                        Username: <span>{userList[userList.length - 1]?.username}</span>
                    </p>
                    <p className="creds">
                        Connection token: <span>{userList[userList.length - 1]?.broker_connection_creds}</span>
                    </p>
                    <p className="note">Please note when you close this modal, you will not be able to restore your user details!!</p>
                </div>
            </Modal>
            <Modal
                header={<DeleteWrapperIcon alt="deleteWrapperIcon" />}
                width="520px"
                height="240px"
                displayButtons={false}
                clickOutside={() => setRemoveUserModalOpen(false)}
                open={removeUserModalOpen}
            >
                <DeleteItemsModal
                    title={userToRemove.revoke ? 'Revoke user' : 'Delete user'}
                    desc={
                        <>
                            Are you sure you want to {userToRemove.revoke ? 'revoke' : 'delete'} <b>{userToRemove.username}</b>?
                        </>
                    }
                    buttontxt={<>I understand, {userToRemove.revoke ? 'revoke' : 'delete'} the user</>}
                    handleDeleteSelected={() => removeUser(userToRemove)}
                    textToConfirm={userToRemove.revoke && 'revoke'}
                    loader={userDeletedLoader}
                />
                <br />
            </Modal>
        </div>
    );
}
export default Users;
